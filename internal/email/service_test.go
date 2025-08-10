package email

import (
	"os"
	"testing"
	"time"

	"ecommerce-website/internal/models"

	"github.com/stretchr/testify/assert"
)

func TestNewService(t *testing.T) {
	tests := []struct {
		name          string
		envVars       map[string]string
		expectEnabled bool
	}{
		{
			name: "all env vars set - enabled",
			envVars: map[string]string{
				"SMTP_HOST":     "smtp.example.com",
				"SMTP_PORT":     "587",
				"SMTP_USERNAME": "user@example.com",
				"SMTP_PASSWORD": "password",
				"FROM_EMAIL":    "noreply@example.com",
			},
			expectEnabled: true,
		},
		{
			name: "missing env vars - disabled",
			envVars: map[string]string{
				"SMTP_HOST": "smtp.example.com",
				// Missing other required vars
			},
			expectEnabled: false,
		},
		{
			name:          "no env vars - disabled",
			envVars:       map[string]string{},
			expectEnabled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			os.Clearenv()

			// Set test environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			service := NewService()

			assert.Equal(t, tt.expectEnabled, service.enabled)

			// Clean up
			os.Clearenv()
		})
	}
}

func TestService_SendOrderStatusUpdate_Disabled(t *testing.T) {
	// Create service with no env vars (disabled)
	os.Clearenv()
	service := NewService()

	order := &models.Order{
		ID:     "test-order-id",
		Status: "shipped",
		User: models.User{
			Email:     "customer@example.com",
			FirstName: "John",
			LastName:  "Doe",
		},
		Total:     99.99,
		CreatedAt: time.Now(),
		Items: []models.OrderItem{
			{
				Product: models.Product{
					Name: "Test Product",
				},
				Quantity: 1,
				Total:    99.99,
			},
		},
		ShippingAddress: models.OrderAddress{
			FirstName:  "John",
			LastName:   "Doe",
			Address1:   "123 Main St",
			City:       "Anytown",
			State:      "CA",
			PostalCode: "12345",
			Country:    "US",
		},
	}

	// Should not return error even when disabled
	err := service.SendOrderStatusUpdate(order, "pending", "shipped")
	assert.NoError(t, err)
}

func TestGetStatusMessage(t *testing.T) {
	tests := []struct {
		status          string
		expectedMessage string
	}{
		{
			status:          "pending",
			expectedMessage: "Your order has been received and is being processed.",
		},
		{
			status:          "processing",
			expectedMessage: "Your order is currently being prepared for shipment.",
		},
		{
			status:          "shipped",
			expectedMessage: "Your order has been shipped and is on its way to you.",
		},
		{
			status:          "delivered",
			expectedMessage: "Your order has been successfully delivered.",
		},
		{
			status:          "cancelled",
			expectedMessage: "Your order has been cancelled.",
		},
		{
			status:          "refunded",
			expectedMessage: "Your order has been refunded.",
		},
		{
			status:          "unknown_status",
			expectedMessage: "Your order status has been updated.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			message := getStatusMessage(tt.status)
			assert.Equal(t, tt.expectedMessage, message)
		})
	}
}

func TestService_SendOrderStatusUpdate_TemplateExecution(t *testing.T) {
	// Set up environment for enabled service
	os.Setenv("SMTP_HOST", "smtp.example.com")
	os.Setenv("SMTP_PORT", "587")
	os.Setenv("SMTP_USERNAME", "user@example.com")
	os.Setenv("SMTP_PASSWORD", "password")
	os.Setenv("FROM_EMAIL", "noreply@example.com")
	defer os.Clearenv()

	service := NewService()

	order := &models.Order{
		ID:     "test-order-id",
		Status: "shipped",
		User: models.User{
			Email:     "customer@example.com",
			FirstName: "John",
			LastName:  "Doe",
		},
		Total:     99.99,
		CreatedAt: time.Now(),
		Items: []models.OrderItem{
			{
				Product: models.Product{
					Name: "Test Product",
				},
				Quantity: 1,
				Total:    99.99,
			},
		},
		ShippingAddress: models.OrderAddress{
			FirstName:  "John",
			LastName:   "Doe",
			Address1:   "123 Main St",
			City:       "Anytown",
			State:      "CA",
			PostalCode: "12345",
			Country:    "US",
		},
	}

	// This will fail at SMTP send, but we can test template execution
	err := service.SendOrderStatusUpdate(order, "pending", "shipped")

	// We expect an error because we're not actually connecting to SMTP
	// but the template should execute successfully
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send email")
}
