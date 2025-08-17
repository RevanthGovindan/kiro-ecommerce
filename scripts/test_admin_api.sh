#!/bin/bash

# Test script for Admin API functionality
BASE_URL="http://localhost:8080"

echo "Testing Admin API functionality..."
echo "Note: Make sure ADMIN_EMAIL and ADMIN_PASSWORD environment variables are set"
echo ""

# Test 1: Admin login with correct credentials
echo "1. Testing Admin Login with correct credentials"
ADMIN_RESPONSE=$(curl -s -X POST "$BASE_URL/api/auth/admin/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@ecommerce.com",
    "password": "admin123456"
  }')

echo "$ADMIN_RESPONSE" | jq '.'

# Extract admin token for further tests
ADMIN_TOKEN=$(echo "$ADMIN_RESPONSE" | jq -r '.data.tokens.access_token // empty')

if [ -z "$ADMIN_TOKEN" ] || [ "$ADMIN_TOKEN" = "null" ]; then
    echo "❌ Failed to get admin token. Cannot proceed with further tests."
    exit 1
fi

echo "✅ Admin token obtained successfully"

# Test 2: Admin login with incorrect credentials
echo -e "\n2. Testing Admin Login with incorrect credentials"
curl -s -X POST "$BASE_URL/api/auth/admin/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@ecommerce.com",
    "password": "wrongpassword"
  }' | jq '.'

# Test 3: Create category with admin token (should succeed)
echo -e "\n3. Testing Create Category with admin token (should succeed)"
curl -s -X POST "$BASE_URL/api/categories" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{
    "name": "Test Admin Category",
    "slug": "test-admin-category",
    "description": "Category created by admin"
  }' | jq '.'

# Test 4: Try to create category without token (should fail)
echo -e "\n4. Testing Create Category without token (should fail with 401)"
curl -s -X POST "$BASE_URL/api/categories" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Unauthorized Category",
    "slug": "unauthorized-category"
  }' | jq '.'

# Test 5: Regular user login and try to create category (should fail with 403)
echo -e "\n5. Testing Create Category with regular user token (should fail with 403)"

# First, try to register a regular user (might fail if user exists, that's ok)
curl -s -X POST "$BASE_URL/api/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@test.com",
    "password": "userpassword123",
    "firstName": "Test",
    "lastName": "User"
  }' > /dev/null

# Login as regular user
USER_RESPONSE=$(curl -s -X POST "$BASE_URL/api/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@test.com",
    "password": "userpassword123"
  }')

USER_TOKEN=$(echo "$USER_RESPONSE" | jq -r '.data.tokens.access_token // empty')

if [ -n "$USER_TOKEN" ] && [ "$USER_TOKEN" != "null" ]; then
    echo "Regular user token obtained, testing category creation..."
    curl -s -X POST "$BASE_URL/api/categories" \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer $USER_TOKEN" \
      -d '{
        "name": "User Category",
        "slug": "user-category"
      }' | jq '.'
else
    echo "Could not obtain regular user token (user might not exist or login failed)"
fi

echo -e "\n" + "="*60
echo "Admin API tests completed!"
echo ""
echo "Expected results:"
echo "✅ Test 1: Admin login should succeed"
echo "❌ Test 2: Admin login with wrong credentials should fail (401)"
echo "✅ Test 3: Category creation with admin token should succeed (201)"
echo "❌ Test 4: Category creation without token should fail (401)"
echo "❌ Test 5: Category creation with user token should fail (403)"