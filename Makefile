# Ecommerce Website Makefile

.PHONY: help dev dev-backend dev-frontend build test clean docker-up docker-down

# Default target
help:
	@echo "Available commands:"
	@echo "  dev           - Start both backend and frontend in development mode"
	@echo "  dev-backend   - Start backend with hot reload"
	@echo "  dev-frontend  - Start frontend development server"
	@echo "  build         - Build both backend and frontend"
	@echo "  test          - Run all tests"
	@echo "  clean         - Clean build artifacts"
	@echo "  docker-up     - Start PostgreSQL and Redis containers"
	@echo "  docker-down   - Stop and remove containers"
	@echo "  db-migrate    - Run database migrations"
	@echo "  db-seed       - Seed database with sample data"
	@echo "  db-reset      - Reset database (drop all tables and recreate with seed data)"

# Development
dev: docker-up
	@echo "Starting development environment..."
	@make -j2 dev-backend dev-frontend

dev-backend:
	@echo "Starting backend with hot reload..."
	air

dev-frontend:
	@echo "Starting frontend development server..."
	cd frontend && npm run dev

# Build
build:
	@echo "Building backend..."
	go build -o bin/server ./cmd/server
	@echo "Building frontend..."
	cd frontend && npm run build

# Testing
test:
	@echo "Running backend tests..."
	go test ./...
	@echo "Running frontend tests..."
	cd frontend && npm test

# Clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -rf tmp/
	rm -rf frontend/.next/
	rm -rf frontend/out/

# Docker
docker-up:
	@echo "Starting PostgreSQL and Redis containers..."
	docker-compose up -d

docker-down:
	@echo "Stopping containers..."
	docker-compose down

# Database
db-migrate:
	@echo "Running database migrations..."
	go run ./cmd/migrate

db-seed:
	@echo "Seeding database..."
	go run ./cmd/migrate -seed

db-reset:
	@echo "Resetting database (dropping all tables and recreating)..."
	go run ./cmd/migrate -drop -seed

db-status:
	@echo "Checking database connection..."
	go run -c 'package main; import ("ecommerce-website/internal/config"; "ecommerce-website/internal/database"; "log"); func main() { cfg := config.Load(); if err := database.Initialize(cfg); err != nil { log.Fatal(err) }; log.Println("Database connection successful"); database.Close() }'