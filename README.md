# Ecommerce Website

A modern ecommerce website built with Go (Gin) backend and Next.js frontend.

## Tech Stack

- **Backend**: Go 1.21+ with Gin framework
- **Frontend**: Next.js 14 with React 18, TypeScript, Tailwind CSS
- **Database**: PostgreSQL 15 with GORM ORM
- **Cache**: Redis 7
- **Authentication**: JWT with refresh tokens
- **Payment**: Stripe integration

## Prerequisites

- Go 1.21 or higher
- Node.js 18 or higher
- Docker and Docker Compose
- Make (optional, for convenience commands)

## Quick Start

1. **Clone and setup**:
   ```bash
   git clone <repository-url>
   cd ecommerce-website
   cp .env.example .env
   ```

2. **Start development environment**:
   ```bash
   # Start PostgreSQL and Redis
   make docker-up
   
   # Start both backend and frontend
   make dev
   ```

   Or manually:
   ```bash
   # Terminal 1: Start containers
   docker-compose up -d
   
   # Terminal 2: Start backend
   air
   
   # Terminal 3: Start frontend
   cd frontend && npm run dev
   ```

3. **Access the application**:
   - Frontend: http://localhost:3000
   - Backend API: http://localhost:8080
   - Health check: http://localhost:8080/health

## Project Structure

```
.
├── cmd/
│   └── server/          # Application entrypoint
├── internal/
│   ├── config/          # Configuration management
│   ├── models/          # Data models
│   ├── handlers/        # HTTP handlers
│   ├── services/        # Business logic
│   ├── repositories/    # Data access layer
│   └── middleware/      # HTTP middleware
├── pkg/
│   └── utils/           # Shared utilities
├── web/                 # Static assets
├── frontend/            # Next.js frontend application
├── docker-compose.yml   # Development containers
└── .air.toml           # Hot reload configuration
```

## Development Commands

```bash
# Backend development
make dev-backend        # Start backend with hot reload
go run cmd/server/main.go  # Run backend directly

# Frontend development
make dev-frontend       # Start frontend dev server
cd frontend && npm run dev  # Run frontend directly

# Database
make docker-up          # Start PostgreSQL and Redis
make docker-down        # Stop containers

# Building
make build              # Build both backend and frontend
go build -o bin/server ./cmd/server  # Build backend only
cd frontend && npm run build  # Build frontend only

# Testing
make test               # Run all tests
go test ./...           # Run backend tests only
cd frontend && npm test # Run frontend tests only
```

## Environment Variables

Copy `.env.example` to `.env` and update the values:

- `PORT`: Server port (default: 8080)
- `DATABASE_URL`: PostgreSQL connection string
- `REDIS_URL`: Redis connection string
- `JWT_SECRET`: Secret key for JWT tokens
- `STRIPE_SECRET_KEY`: Stripe secret key for payments

## API Endpoints

### Health Check
- `GET /health` - API health status

### Authentication (Coming Soon)
- `POST /api/auth/register` - User registration
- `POST /api/auth/login` - User login
- `POST /api/auth/logout` - User logout

### Products (Coming Soon)
- `GET /api/products` - List products
- `GET /api/products/:id` - Get product details

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `make test`
5. Submit a pull request

## License

This project is licensed under the MIT License.