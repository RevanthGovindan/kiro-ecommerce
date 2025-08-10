#!/bin/bash

# Test script for orders API endpoints
BASE_URL="http://localhost:8080"

echo "🛒 Testing Orders API..."
echo "================================"

# First, register and login to get a token
echo "1. Setting up test user..."
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/api/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "ordertest@example.com",
    "password": "testpassword123",
    "firstName": "Order",
    "lastName": "Tester"
  }')

LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/api/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "ordertest@example.com",
    "password": "testpassword123"
  }')

ACCESS_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.data.tokens.access_token')
echo "Access token obtained: ${ACCESS_TOKEN:0:20}..."
echo ""

if [ "$ACCESS_TOKEN" != "null" ] && [ "$ACCESS_TOKEN" != "" ]; then
  
  # Test getting user orders (should be empty initially)
  echo "2. Testing GET /api/orders (user orders)..."
  curl -s -X GET "$BASE_URL/api/orders" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq '.'
  echo ""

  # Test creating an order (should fail with empty cart)
  echo "3. Testing POST /api/orders/create (should fail with empty cart)..."
  CREATE_ORDER_RESPONSE=$(curl -s -X POST "$BASE_URL/api/orders/create" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
      "sessionId": "test-session-123",
      "shippingAddress": {
        "firstName": "John",
        "lastName": "Doe",
        "address1": "123 Main St",
        "city": "Anytown",
        "state": "CA",
        "postalCode": "12345",
        "country": "US"
      },
      "billingAddress": {
        "firstName": "John",
        "lastName": "Doe",
        "address1": "123 Main St",
        "city": "Anytown",
        "state": "CA",
        "postalCode": "12345",
        "country": "US"
      },
      "paymentIntentId": "pi_test123"
    }')
  
  echo "$CREATE_ORDER_RESPONSE" | jq '.'
  echo ""

  # Test getting a non-existent order
  echo "4. Testing GET /api/orders/non-existent-id..."
  curl -s -X GET "$BASE_URL/api/orders/non-existent-id" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq '.'
  echo ""

  # Test without authentication
  echo "5. Testing GET /api/orders without authentication..."
  curl -s -X GET "$BASE_URL/api/orders" | jq '.'
  echo ""

else
  echo "❌ Failed to get access token, skipping order tests"
fi

echo "✅ Orders API testing completed!"