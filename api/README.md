# Ecommerce Website API Documentation

This directory contains the OpenAPI specification for the Ecommerce Website API.

## Files

- `openapi.yaml` - Complete OpenAPI 3.0.3 specification for all API endpoints

## API Overview

The Ecommerce Website API provides the following functionality:

### 🔐 Authentication & Authorization
- User registration and login
- JWT token-based authentication
- Password reset functionality
- Email verification
- Token refresh mechanism

### 👤 User Profile Management
- View and update user profile
- Address management (shipping/billing)
- Order history access

### 🛍️ Product Catalog
- Browse products with filtering and pagination
- Search products
- View product details
- Category management

### 🛒 Shopping Cart
- Add/remove items from cart
- Update item quantities
- Cart persistence across sessions

## Using the OpenAPI Specification

### 1. View Documentation

You can view the API documentation using various tools:

#### Swagger UI (Online)
1. Go to [Swagger Editor](https://editor.swagger.io/)
2. Copy the contents of `openapi.yaml` and paste it into the editor
3. The documentation will be rendered with an interactive interface

#### Local Swagger UI
```bash
# Using Docker
docker run -p 8080:8080 -e SWAGGER_JSON=/api/openapi.yaml -v $(pwd)/api:/api swaggerapi/swagger-ui

# Then open http://localhost:8080 in your browser
```

#### Redoc
```bash
# Using npx
npx redoc-cli serve api/openapi.yaml

# Then open http://localhost:8080 in your browser
```

### 2. Generate Client SDKs

You can generate client SDKs in various programming languages:

```bash
# Install OpenAPI Generator
npm install @openapitools/openapi-generator-cli -g

# Generate JavaScript/TypeScript client
openapi-generator-cli generate -i api/openapi.yaml -g typescript-axios -o clients/typescript

# Generate Python client
openapi-generator-cli generate -i api/openapi.yaml -g python -o clients/python

# Generate Go client
openapi-generator-cli generate -i api/openapi.yaml -g go -o clients/go

# Generate Java client
openapi-generator-cli generate -i api/openapi.yaml -g java -o clients/java
```

### 3. API Testing

#### Using curl
```bash
# Health check
curl -X GET "http://localhost:8080/health"

# Register a new user
curl -X POST "http://localhost:8080/api/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securepassword123",
    "firstName": "John",
    "lastName": "Doe"
  }'

# Login
curl -X POST "http://localhost:8080/api/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securepassword123"
  }'

# Get products (with authentication)
curl -X GET "http://localhost:8080/api/products" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

#### Using Postman
1. Import the `openapi.yaml` file into Postman
2. Postman will automatically create a collection with all endpoints
3. Set up environment variables for base URL and authentication tokens

#### Using Insomnia
1. Import the `openapi.yaml` file into Insomnia
2. All endpoints will be available in your workspace
3. Configure authentication using the Bearer token method

### 4. API Validation

You can validate the OpenAPI specification:

```bash
# Using swagger-codegen
swagger-codegen validate -i api/openapi.yaml

# Using openapi-generator
openapi-generator-cli validate -i api/openapi.yaml

# Using spectral (advanced linting)
npm install -g @stoplight/spectral-cli
spectral lint api/openapi.yaml
```

## API Endpoints Summary

### Authentication (`/api/auth`)
- `POST /register` - Register new user
- `POST /login` - User login
- `POST /refresh` - Refresh JWT token
- `POST /logout` - User logout
- `GET /me` - Get current user info
- `POST /forgot-password` - Request password reset
- `POST /reset-password` - Reset password
- `GET /verify-email` - Verify email address
- `POST /resend-verification` - Resend verification email

### Products (`/api/products`)
- `GET /` - List products with filtering/pagination
- `GET /search` - Search products
- `GET /{id}` - Get product by ID

### Categories (`/api/categories`)
- `GET /` - List all categories
- `GET /{id}` - Get category by ID

### User Profile (`/api/users`)
- `GET /profile` - Get user profile
- `PUT /profile` - Update user profile
- `GET /orders` - Get user order history

### Address Management (`/api/users/addresses`)
- `GET /` - List user addresses
- `POST /` - Create new address
- `GET /{id}` - Get address by ID
- `PUT /{id}` - Update address
- `DELETE /{id}` - Delete address

### Shopping Cart (`/api/cart`)
- `GET /` - Get current cart
- `POST /items` - Add item to cart
- `PUT /items` - Update cart item
- `DELETE /items` - Remove item from cart
- `DELETE /clear` - Clear entire cart

## Response Format

All API responses follow a consistent format:

### Success Response
```json
{
  "success": true,
  "message": "Operation completed successfully",
  "data": { /* response data */ },
  "timestamp": "2024-01-01T12:00:00Z"
}
```

### Error Response
```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable error message",
    "details": { /* additional error details */ }
  },
  "timestamp": "2024-01-01T12:00:00Z"
}
```

## Authentication

Most endpoints require JWT authentication. Include the JWT token in the Authorization header:

```
Authorization: Bearer <your-jwt-token>
```

Tokens are obtained through the login endpoint and can be refreshed using the refresh endpoint.

## Rate Limiting

The API implements rate limiting to prevent abuse. Current limits:
- 100 requests per minute per IP for unauthenticated endpoints
- 1000 requests per minute per user for authenticated endpoints

## Error Codes

Common error codes used throughout the API:

- `VALIDATION_ERROR` - Request validation failed
- `UNAUTHORIZED` - Authentication required
- `FORBIDDEN` - Insufficient permissions
- `NOT_FOUND` - Resource not found
- `CONFLICT` - Resource already exists
- `INTERNAL_ERROR` - Server error

## Support

For API support and questions:
- Email: support@ecommerce.com
- Documentation: This OpenAPI specification
- Issues: Create an issue in the project repository