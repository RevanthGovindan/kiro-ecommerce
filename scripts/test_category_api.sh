#!/bin/bash

# Test script for Category API
BASE_URL="http://localhost:8080"

echo "Testing Category API..."

# Test 1: Get all categories (should work without auth)
echo "1. Testing GET /api/categories"
curl -s -X GET "$BASE_URL/api/categories" | jq '.'

echo -e "\n2. Testing GET /api/categories/non-existent (should return 404)"
curl -s -X GET "$BASE_URL/api/categories/non-existent" | jq '.'

# Test 3: Try to create category without auth (should fail)
echo -e "\n3. Testing POST /api/categories without auth (should return 401)"
curl -s -X POST "$BASE_URL/api/categories" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Electronics",
    "slug": "electronics",
    "description": "Electronic devices and accessories"
  }' | jq '.'

echo -e "\nCategory API tests completed!"