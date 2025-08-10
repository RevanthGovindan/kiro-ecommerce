# Razorpay Payment Integration

This package provides Razorpay payment integration for the ecommerce website.

## Features

- Create Razorpay payment orders
- Verify payment signatures
- Handle payment webhooks
- Payment status tracking
- Comprehensive error handling and logging

## Environment Variables

Set the following environment variables:

```bash
RAZORPAY_KEY_ID=your_razorpay_key_id
RAZORPAY_SECRET=your_razorpay_secret
```

## API Endpoints

### 1. Create Payment Order

**POST** `/api/payments/create-order`

Creates a new Razorpay order for payment processing.

**Headers:**
- `Authorization: Bearer <token>` (required)
- `Content-Type: application/json`

**Request Body:**
```json
{
  "orderId": "order-uuid",
  "amount": 110.50,
  "currency": "INR",
  "description": "Payment for order #123"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Payment order created successfully",
  "data": {
    "id": "payment-uuid",
    "razorpay_order_id": "order_razorpay_id",
    "amount": 11050,
    "currency": "INR",
    "status": "created"
  },
  "timestamp": "2025-01-08T10:30:00Z"
}
```

### 2. Verify Payment

**POST** `/api/payments/verify`

Verifies the payment signature after successful payment.

**Headers:**
- `Authorization: Bearer <token>` (required)
- `Content-Type: application/json`

**Request Body:**
```json
{
  "razorpay_order_id": "order_razorpay_id",
  "razorpay_payment_id": "pay_razorpay_id",
  "razorpay_signature": "signature_hash"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Payment verified successfully",
  "timestamp": "2025-01-08T10:35:00Z"
}
```

### 3. Get Payment Status

**GET** `/api/payments/status/:orderId`

Retrieves the payment status for a specific order.

**Headers:**
- `Authorization: Bearer <token>` (required)

**Response:**
```json
{
  "success": true,
  "message": "Payment status retrieved successfully",
  "data": {
    "id": "payment-uuid",
    "orderId": "order-uuid",
    "razorpayOrderId": "order_razorpay_id",
    "razorpayPaymentId": "pay_razorpay_id",
    "amount": 11050,
    "currency": "INR",
    "status": "paid",
    "method": "card",
    "createdAt": "2025-01-08T10:30:00Z",
    "updatedAt": "2025-01-08T10:35:00Z"
  },
  "timestamp": "2025-01-08T10:40:00Z"
}
```

### 4. Webhook Handler

**POST** `/api/payments/webhook`

Handles Razorpay webhook events for payment status updates.

**Headers:**
- `Content-Type: application/json`

**Request Body:** (Razorpay webhook payload)

**Response:**
```json
{
  "status": "ok"
}
```

## Payment Flow

1. **Create Order**: Frontend calls `/api/payments/create-order` to create a Razorpay order
2. **Payment UI**: Frontend uses Razorpay Checkout with the returned order ID
3. **Verify Payment**: After successful payment, frontend calls `/api/payments/verify` with payment details
4. **Webhook Processing**: Razorpay sends webhook events for payment status updates
5. **Order Update**: Payment verification updates the order status to "paid"

## Frontend Integration Example

```javascript
// 1. Create payment order
const createOrder = async (orderData) => {
  const response = await fetch('/api/payments/create-order', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    },
    body: JSON.stringify(orderData)
  });
  return response.json();
};

// 2. Initialize Razorpay Checkout
const initializePayment = (paymentOrder) => {
  const options = {
    key: 'your_razorpay_key_id',
    amount: paymentOrder.amount,
    currency: paymentOrder.currency,
    order_id: paymentOrder.razorpay_order_id,
    name: 'Your Store Name',
    description: 'Payment for your order',
    handler: async (response) => {
      // 3. Verify payment
      await verifyPayment(response);
    },
    prefill: {
      name: 'Customer Name',
      email: 'customer@example.com',
      contact: '9999999999'
    }
  };
  
  const rzp = new Razorpay(options);
  rzp.open();
};

// 3. Verify payment
const verifyPayment = async (paymentResponse) => {
  const response = await fetch('/api/payments/verify', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    },
    body: JSON.stringify({
      razorpay_order_id: paymentResponse.razorpay_order_id,
      razorpay_payment_id: paymentResponse.razorpay_payment_id,
      razorpay_signature: paymentResponse.razorpay_signature
    })
  });
  
  if (response.ok) {
    // Payment verified successfully
    window.location.href = '/order-success';
  } else {
    // Payment verification failed
    alert('Payment verification failed');
  }
};
```

## Error Codes

- `INVALID_REQUEST`: Invalid request format or missing required fields
- `PAYMENT_ORDER_FAILED`: Failed to create Razorpay order
- `PAYMENT_VERIFICATION_FAILED`: Payment signature verification failed
- `PAYMENT_NOT_FOUND`: Payment record not found
- `WEBHOOK_PROCESSING_FAILED`: Failed to process webhook event
- `MISSING_ORDER_ID`: Order ID parameter is missing
- `INVALID_WEBHOOK_PAYLOAD`: Invalid webhook payload format

## Testing

Run the payment tests:

```bash
# Run all payment tests
go test ./internal/payments/... -v

# Run only integration tests
go test ./internal/payments/... -v -run Integration

# Run specific test
go test ./internal/payments/... -v -run TestService_CreateOrder
```

## Security Considerations

1. **Signature Verification**: All payments are verified using HMAC-SHA256 signature
2. **Authentication**: All payment endpoints (except webhooks) require valid JWT tokens
3. **Environment Variables**: Razorpay credentials are stored as environment variables
4. **Webhook Security**: Webhook endpoint should be configured with proper signature verification in production

## Database Schema

The payment integration uses the following database table:

```sql
CREATE TABLE payments (
    id VARCHAR(36) PRIMARY KEY,
    order_id VARCHAR(36) NOT NULL,
    razorpay_order_id VARCHAR(255) UNIQUE NOT NULL,
    razorpay_payment_id VARCHAR(255) UNIQUE,
    razorpay_signature VARCHAR(255),
    amount BIGINT NOT NULL,
    currency VARCHAR(3) DEFAULT 'INR',
    status VARCHAR(20) DEFAULT 'created',
    method VARCHAR(50),
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_payments_order_id (order_id),
    INDEX idx_payments_razorpay_order_id (razorpay_order_id),
    INDEX idx_payments_status (status)
);
```

## Supported Payment Methods

Razorpay supports various payment methods:
- Credit/Debit Cards
- Net Banking
- UPI
- Wallets (Paytm, PhonePe, etc.)
- EMI
- Bank Transfers

The payment method used will be automatically detected and stored in the `method` field after successful payment.