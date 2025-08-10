package payments

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	"ecommerce-website/internal/models"

	"github.com/razorpay/razorpay-go"
	"gorm.io/gorm"
)

type Service struct {
	db     *gorm.DB
	client *razorpay.Client
	secret string
}

type CreateOrderRequest struct {
	OrderID     string  `json:"orderId" binding:"required"`
	Amount      float64 `json:"amount" binding:"required"`
	Currency    string  `json:"currency"`
	Description string  `json:"description"`
}

type VerifyPaymentRequest struct {
	RazorpayOrderID   string `json:"razorpay_order_id" binding:"required"`
	RazorpayPaymentID string `json:"razorpay_payment_id" binding:"required"`
	RazorpaySignature string `json:"razorpay_signature" binding:"required"`
}

type PaymentResponse struct {
	ID              string `json:"id"`
	RazorpayOrderID string `json:"razorpay_order_id"`
	Amount          int64  `json:"amount"`
	Currency        string `json:"currency"`
	Status          string `json:"status"`
}

func NewService(db *gorm.DB, keyID, keySecret string) *Service {
	client := razorpay.NewClient(keyID, keySecret)
	return &Service{
		db:     db,
		client: client,
		secret: keySecret,
	}
}

func (s *Service) CreateOrder(req CreateOrderRequest) (*PaymentResponse, error) {
	// Verify the order exists and get its details
	var order models.Order
	if err := s.db.First(&order, "id = ?", req.OrderID).Error; err != nil {
		return nil, fmt.Errorf("order not found: %w", err)
	}

	// Convert amount to paise (Razorpay expects amount in smallest currency unit)
	amountInPaise := int64(req.Amount * 100)

	// Create Razorpay order
	data := map[string]interface{}{
		"amount":   amountInPaise,
		"currency": "INR",
		"receipt":  req.OrderID,
	}

	if req.Description != "" {
		data["notes"] = map[string]interface{}{
			"description": req.Description,
		}
	}

	razorpayOrder, err := s.client.Order.Create(data, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Razorpay order: %w", err)
	}

	// Save payment record in database
	payment := models.Payment{
		OrderID:         req.OrderID,
		RazorpayOrderID: razorpayOrder["id"].(string),
		Amount:          amountInPaise,
		Currency:        "INR",
		Status:          models.PaymentStatusCreated,
		Description:     &req.Description,
	}

	if err := s.db.Create(&payment).Error; err != nil {
		return nil, fmt.Errorf("failed to save payment record: %w", err)
	}

	return &PaymentResponse{
		ID:              payment.ID,
		RazorpayOrderID: payment.RazorpayOrderID,
		Amount:          payment.Amount,
		Currency:        payment.Currency,
		Status:          payment.Status,
	}, nil
}

func (s *Service) VerifyPayment(req VerifyPaymentRequest) error {
	// Verify signature
	if !s.verifySignature(req.RazorpayOrderID, req.RazorpayPaymentID, req.RazorpaySignature) {
		return errors.New("invalid payment signature")
	}

	// Find payment record
	var payment models.Payment
	if err := s.db.First(&payment, "razorpay_order_id = ?", req.RazorpayOrderID).Error; err != nil {
		return fmt.Errorf("payment record not found: %w", err)
	}

	// Update payment record
	payment.RazorpayPaymentID = &req.RazorpayPaymentID
	payment.RazorpaySignature = &req.RazorpaySignature
	payment.Status = models.PaymentStatusPaid

	if err := s.db.Save(&payment).Error; err != nil {
		return fmt.Errorf("failed to update payment record: %w", err)
	}

	// Update order status
	if err := s.db.Model(&models.Order{}).Where("id = ?", payment.OrderID).Update("status", "paid").Error; err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	return nil
}

func (s *Service) GetPaymentByOrderID(orderID string) (*models.Payment, error) {
	var payment models.Payment
	if err := s.db.First(&payment, "order_id = ?", orderID).Error; err != nil {
		return nil, err
	}
	return &payment, nil
}

func (s *Service) verifySignature(orderID, paymentID, signature string) bool {
	// Create the expected signature
	message := orderID + "|" + paymentID
	h := hmac.New(sha256.New, []byte(s.secret))
	h.Write([]byte(message))
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

func (s *Service) HandleWebhook(payload map[string]interface{}) error {
	event, ok := payload["event"].(string)
	if !ok {
		return errors.New("invalid webhook payload: missing event")
	}

	switch event {
	case "payment.captured":
		return s.handlePaymentCaptured(payload)
	case "payment.failed":
		return s.handlePaymentFailed(payload)
	default:
		// Ignore other events
		return nil
	}
}

func (s *Service) handlePaymentCaptured(payload map[string]interface{}) error {
	paymentData, ok := payload["payload"].(map[string]interface{})
	if !ok {
		return errors.New("invalid webhook payload: missing payment data")
	}

	payment, ok := paymentData["payment"].(map[string]interface{})
	if !ok {
		return errors.New("invalid webhook payload: missing payment entity")
	}

	orderID, ok := payment["order_id"].(string)
	if !ok {
		return errors.New("invalid webhook payload: missing order_id")
	}

	paymentID, ok := payment["id"].(string)
	if !ok {
		return errors.New("invalid webhook payload: missing payment id")
	}

	// Update payment record
	var paymentRecord models.Payment
	if err := s.db.First(&paymentRecord, "razorpay_order_id = ?", orderID).Error; err != nil {
		return fmt.Errorf("payment record not found: %w", err)
	}

	paymentRecord.RazorpayPaymentID = &paymentID
	paymentRecord.Status = models.PaymentStatusPaid

	if method, ok := payment["method"].(string); ok {
		paymentRecord.Method = &method
	}

	if err := s.db.Save(&paymentRecord).Error; err != nil {
		return fmt.Errorf("failed to update payment record: %w", err)
	}

	// Update order status
	if err := s.db.Model(&models.Order{}).Where("id = ?", paymentRecord.OrderID).Update("status", "paid").Error; err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	return nil
}

func (s *Service) handlePaymentFailed(payload map[string]interface{}) error {
	paymentData, ok := payload["payload"].(map[string]interface{})
	if !ok {
		return errors.New("invalid webhook payload: missing payment data")
	}

	payment, ok := paymentData["payment"].(map[string]interface{})
	if !ok {
		return errors.New("invalid webhook payload: missing payment entity")
	}

	orderID, ok := payment["order_id"].(string)
	if !ok {
		return errors.New("invalid webhook payload: missing order_id")
	}

	// Update payment record
	var paymentRecord models.Payment
	if err := s.db.First(&paymentRecord, "razorpay_order_id = ?", orderID).Error; err != nil {
		return fmt.Errorf("payment record not found: %w", err)
	}

	paymentRecord.Status = models.PaymentStatusFailed

	if err := s.db.Save(&paymentRecord).Error; err != nil {
		return fmt.Errorf("failed to update payment record: %w", err)
	}

	// Update order status
	if err := s.db.Model(&models.Order{}).Where("id = ?", paymentRecord.OrderID).Update("status", "payment_failed").Error; err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	return nil
}
