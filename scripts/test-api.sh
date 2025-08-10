#!/bin/bash

# Simple API test script
# This script tests basic API endpoints to ensure they're working

BASE_URL="http://localhost:8080"

echo "🧪 Testing Ecommerce API..."
echo "================================"

# Test health endpoint
echo "1. Testing health endpoint..."
curl -s "$BASE_URL/health" | jq '.'
echo ""

# Test user registration
echo "2. Testing user registration..."
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/api/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "testpassword123",
    "firstName": "John",
    "lastName": "Doe",
    "phone": "+1234567890"
  }')

echo "$REGISTER_RESPONSE" | jq '.'
echo ""

# Test user login
echo "3. Testing user login..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/api/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "testpassword123"
  }')

echo "$LOGIN_RESPONSE" | jq '.'

# Extract access token for authenticated requests
ACCESS_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.data.tokens.access_token')
echo ""

if [ "$ACCESS_TOKEN" != "null" ] && [ "$ACCESS_TOKEN" != "" ]; then
  echo "4. Testing authenticated endpoint (user profile)..."
  curl -s -X GET "$BASE_URL/api/users/profile" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq '.'
  echo ""

  echo "5. Testing products endpoint..."
  curl -s -X GET "$BASE_URL/api/products" | jq '.'
  echo ""

  echo "6. Testing categories endpoint..."
  curl -s -X GET "$BASE_URL/api/categories" | jq '.'
  echo ""

  echo "7. Testing cart endpoint..."
  curl -s -X GET "$BASE_URL/api/cart" | jq '.'
  echo ""
else
  echo "❌ Login failed, skipping authenticated tests"
fi

echo "✅ API testing completed!"