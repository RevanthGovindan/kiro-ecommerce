package email

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/smtp"
	"os"
	"strings"

	"ecommerce-website/internal/models"
)

// ServiceInterface defines the interface for email service
type ServiceInterface interface {
	SendOrderStatusUpdate(order *models.Order, oldStatus, newStatus string) error
}

type Service struct {
	smtpHost     string
	smtpPort     string
	smtpUsername string
	smtpPassword string
	fromEmail    string
	enabled      bool
}

// NewService creates a new email service
func NewService() *Service {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUsername := os.Getenv("SMTP_USERNAME")
	smtpPassword := os.Getenv("SMTP_PASSWORD")
	fromEmail := os.Getenv("FROM_EMAIL")

	// Email service is enabled only if all required env vars are set
	enabled := smtpHost != "" && smtpPort != "" && smtpUsername != "" && smtpPassword != "" && fromEmail != ""

	if !enabled {
		log.Println("Email service disabled: missing SMTP configuration")
	}

	return &Service{
		smtpHost:     smtpHost,
		smtpPort:     smtpPort,
		smtpUsername: smtpUsername,
		smtpPassword: smtpPassword,
		fromEmail:    fromEmail,
		enabled:      enabled,
	}
}

// SendOrderStatusUpdate sends an email notification when order status changes
func (s *Service) SendOrderStatusUpdate(order *models.Order, oldStatus, newStatus string) error {
	if !s.enabled {
		log.Printf("Email service disabled, skipping order status update notification for order %s", order.ID)
		return nil
	}

	// Prepare email data
	data := struct {
		Order         *models.Order
		OldStatus     string
		NewStatus     string
		StatusMessage string
	}{
		Order:         order,
		OldStatus:     oldStatus,
		NewStatus:     newStatus,
		StatusMessage: getStatusMessage(newStatus),
	}

	// Parse email template with custom functions
	tmpl, err := template.New("order_status_update").Funcs(template.FuncMap{
		"title": func(s string) string {
			if len(s) == 0 {
				return s
			}
			return strings.ToUpper(s[:1]) + s[1:]
		},
	}).Parse(orderStatusUpdateTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse email template: %w", err)
	}

	// Execute template
	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to execute email template: %w", err)
	}

	// Prepare email message
	subject := fmt.Sprintf("Order Update - Order #%s", order.ID[:8])
	message := fmt.Sprintf("From: %s\r\n", s.fromEmail) +
		fmt.Sprintf("To: %s\r\n", order.User.Email) +
		fmt.Sprintf("Subject: %s\r\n", subject) +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n" +
		"\r\n" +
		body.String()

	// Send email
	auth := smtp.PlainAuth("", s.smtpUsername, s.smtpPassword, s.smtpHost)
	addr := fmt.Sprintf("%s:%s", s.smtpHost, s.smtpPort)

	err = smtp.SendMail(addr, auth, s.fromEmail, []string{order.User.Email}, []byte(message))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Printf("Order status update email sent to %s for order %s", order.User.Email, order.ID)
	return nil
}

// getStatusMessage returns a user-friendly message for each order status
func getStatusMessage(status string) string {
	messages := map[string]string{
		"pending":    "Your order has been received and is being processed.",
		"processing": "Your order is currently being prepared for shipment.",
		"shipped":    "Your order has been shipped and is on its way to you.",
		"delivered":  "Your order has been successfully delivered.",
		"cancelled":  "Your order has been cancelled.",
		"refunded":   "Your order has been refunded.",
	}

	if message, exists := messages[status]; exists {
		return message
	}
	return "Your order status has been updated."
}

// Email template for order status updates
const orderStatusUpdateTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Order Update</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #f8f9fa; padding: 20px; text-align: center; }
        .content { padding: 20px; }
        .order-details { background-color: #f8f9fa; padding: 15px; margin: 20px 0; }
        .status-update { background-color: #d4edda; border: 1px solid #c3e6cb; padding: 15px; margin: 20px 0; border-radius: 5px; }
        .footer { text-align: center; padding: 20px; color: #666; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Order Status Update</h1>
        </div>
        
        <div class="content">
            <p>Hello {{.Order.User.FirstName}} {{.Order.User.LastName}},</p>
            
            <div class="status-update">
                <h3>Your order status has been updated!</h3>
                <p><strong>Order ID:</strong> #{{.Order.ID}}</p>
                <p><strong>New Status:</strong> {{.NewStatus | title}}</p>
                <p>{{.StatusMessage}}</p>
            </div>
            
            <div class="order-details">
                <h3>Order Details</h3>
                <p><strong>Order Total:</strong> ${{printf "%.2f" .Order.Total}}</p>
                <p><strong>Order Date:</strong> {{.Order.CreatedAt.Format "January 2, 2006"}}</p>
                
                <h4>Items:</h4>
                {{range .Order.Items}}
                <p>• {{.Product.Name}} (Qty: {{.Quantity}}) - ${{printf "%.2f" .Total}}</p>
                {{end}}
                
                <h4>Shipping Address:</h4>
                <p>
                    {{.Order.ShippingAddress.FirstName}} {{.Order.ShippingAddress.LastName}}<br>
                    {{.Order.ShippingAddress.Address1}}<br>
                    {{if .Order.ShippingAddress.Address2}}{{.Order.ShippingAddress.Address2}}<br>{{end}}
                    {{.Order.ShippingAddress.City}}, {{.Order.ShippingAddress.State}} {{.Order.ShippingAddress.PostalCode}}<br>
                    {{.Order.ShippingAddress.Country}}
                </p>
            </div>
            
            <p>If you have any questions about your order, please don't hesitate to contact our customer support team.</p>
            
            <p>Thank you for your business!</p>
        </div>
        
        <div class="footer">
            <p>This is an automated message. Please do not reply to this email.</p>
        </div>
    </div>
</body>
</html>
`
