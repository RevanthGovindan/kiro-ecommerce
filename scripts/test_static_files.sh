#!/bin/bash

# Test script for static file serving in development mode
BASE_URL="http://localhost:8080"

echo "Testing Static File Serving in Development Mode..."
echo "Note: Make sure ENVIRONMENT=development is set and server is running"
echo ""

# Test 1: Check if root redirects to /api-docs/
echo "1. Testing GET / (should redirect to /api-docs/)"
curl -s -I "$BASE_URL/" | grep -E "(Location|HTTP)"

# Test 2: Check if /docs redirects to /api-docs/
echo -e "\n2. Testing GET /docs (should redirect to /api-docs/)"
curl -s -I "$BASE_URL/docs" | grep -E "(Location|HTTP)"

# Test 3: Check if /api-docs/ serves the HTML file
echo -e "\n3. Testing GET /api-docs/ (should serve index.html)"
response=$(curl -s "$BASE_URL/api-docs/")
if echo "$response" | grep -q "rapi-doc"; then
    echo "✅ HTML file served successfully (contains rapi-doc)"
else
    echo "❌ HTML file not served correctly"
fi

# Test 4: Check if /api-docs/openapi.yaml serves the OpenAPI spec
echo -e "\n4. Testing GET /api-docs/openapi.yaml (should serve OpenAPI spec)"
response=$(curl -s "$BASE_URL/api-docs/openapi.yaml")
if echo "$response" | grep -q "openapi: 3.0.3"; then
    echo "✅ OpenAPI spec served successfully"
else
    echo "❌ OpenAPI spec not served correctly"
fi

# Test 5: Check if /api-docs/README.md serves the README
echo -e "\n5. Testing GET /api-docs/README.md (should serve README)"
response=$(curl -s "$BASE_URL/api-docs/README.md")
if [ -n "$response" ]; then
    echo "✅ README.md served successfully"
else
    echo "❌ README.md not served correctly"
fi

echo -e "\n" + "="*50
echo "Static file serving tests completed!"
echo ""
echo "If running in development mode, you can access:"
echo "🌐 Root redirect: http://localhost:8080/"
echo "🌐 Docs redirect: http://localhost:8080/docs"
echo "📚 API Documentation: http://localhost:8080/api-docs/"
echo "📄 OpenAPI Spec: http://localhost:8080/api-docs/openapi.yaml"
echo "📖 README: http://localhost:8080/api-docs/README.md"