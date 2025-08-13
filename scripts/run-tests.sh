#!/bin/bash

# Comprehensive test runner script for the ecommerce website

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BACKEND_PORT=8080
FRONTEND_PORT=3000
TEST_DB_NAME="ecommerce_test"

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if a service is running
check_service() {
    local port=$1
    local service_name=$2
    
    if curl -s "http://localhost:$port" > /dev/null 2>&1; then
        print_success "$service_name is running on port $port"
        return 0
    else
        print_warning "$service_name is not running on port $port"
        return 1
    fi
}

# Function to wait for service to be ready
wait_for_service() {
    local port=$1
    local service_name=$2
    local max_attempts=30
    local attempt=1
    
    print_status "Waiting for $service_name to be ready..."
    
    while [ $attempt -le $max_attempts ]; do
        if check_service $port "$service_name"; then
            return 0
        fi
        
        print_status "Attempt $attempt/$max_attempts - waiting for $service_name..."
        sleep 2
        attempt=$((attempt + 1))
    done
    
    print_error "$service_name failed to start within expected time"
    return 1
}

# Function to run backend tests
run_backend_tests() {
    print_status "Running backend tests..."
    
    # Set test environment variables
    export DATABASE_URL="postgres://testuser:testpass@localhost:5432/${TEST_DB_NAME}?sslmode=disable"
    export REDIS_URL="redis://localhost:6379"
    export JWT_SECRET="test-secret-key"
    export ENVIRONMENT="test"
    
    # Run unit tests
    print_status "Running Go unit tests..."
    go test -v -race -coverprofile=coverage.out ./internal/...
    
    if [ $? -eq 0 ]; then
        print_success "Backend unit tests passed"
    else
        print_error "Backend unit tests failed"
        return 1
    fi
    
    # Run integration tests
    print_status "Running Go integration tests..."
    go test -v -tags=integration ./tests/integration/...
    
    if [ $? -eq 0 ]; then
        print_success "Backend integration tests passed"
    else
        print_error "Backend integration tests failed"
        return 1
    fi
    
    # Generate coverage report
    go tool cover -html=coverage.out -o coverage.html
    print_success "Coverage report generated: coverage.html"
}

# Function to run frontend tests
run_frontend_tests() {
    print_status "Running frontend tests..."
    
    cd frontend
    
    # Run unit tests
    print_status "Running Jest unit tests..."
    npm run test -- --coverage --watchAll=false
    
    if [ $? -eq 0 ]; then
        print_success "Frontend unit tests passed"
    else
        print_error "Frontend unit tests failed"
        cd ..
        return 1
    fi
    
    cd ..
}

# Function to run E2E tests
run_e2e_tests() {
    print_status "Running E2E tests..."
    
    # Check if services are running
    if ! check_service $BACKEND_PORT "Backend API"; then
        print_error "Backend API must be running for E2E tests"
        return 1
    fi
    
    if ! check_service $FRONTEND_PORT "Frontend"; then
        print_error "Frontend must be running for E2E tests"
        return 1
    fi
    
    cd frontend
    
    # Install Playwright browsers if needed
    npx playwright install --with-deps
    
    # Run E2E tests
    print_status "Running Playwright E2E tests..."
    npm run test:e2e
    
    if [ $? -eq 0 ]; then
        print_success "E2E tests passed"
    else
        print_error "E2E tests failed"
        cd ..
        return 1
    fi
    
    cd ..
}

# Function to run performance tests
run_performance_tests() {
    print_status "Running performance tests..."
    
    # Check if backend is running
    if ! check_service $BACKEND_PORT "Backend API"; then
        print_error "Backend API must be running for performance tests"
        return 1
    fi
    
    # Check if k6 is installed
    if ! command -v k6 &> /dev/null; then
        print_error "k6 is not installed. Please install k6 to run performance tests."
        return 1
    fi
    
    # Run load tests
    print_status "Running load tests..."
    k6 run tests/performance/load-test.js
    
    if [ $? -eq 0 ]; then
        print_success "Load tests completed"
    else
        print_error "Load tests failed"
        return 1
    fi
}

# Function to setup test environment
setup_test_env() {
    print_status "Setting up test environment..."
    
    # Check if Docker is running
    if ! docker info > /dev/null 2>&1; then
        print_error "Docker is not running. Please start Docker."
        return 1
    fi
    
    # Start test databases with docker-compose
    print_status "Starting test databases..."
    docker-compose -f docker-compose.test.yml up -d
    
    # Wait for databases to be ready
    sleep 10
    
    # Run database migrations
    print_status "Running database migrations..."
    go run cmd/migrate/main.go
    
    print_success "Test environment setup complete"
}

# Function to cleanup test environment
cleanup_test_env() {
    print_status "Cleaning up test environment..."
    
    # Stop test databases
    docker-compose -f docker-compose.test.yml down
    
    # Remove test coverage files
    rm -f coverage.out coverage.html
    
    print_success "Test environment cleanup complete"
}

# Function to show usage
show_usage() {
    echo "Usage: $0 [OPTIONS] [TEST_TYPE]"
    echo ""
    echo "Test Types:"
    echo "  unit        Run unit tests only"
    echo "  integration Run integration tests only"
    echo "  e2e         Run E2E tests only"
    echo "  performance Run performance tests only"
    echo "  all         Run all tests (default)"
    echo ""
    echo "Options:"
    echo "  --setup     Setup test environment before running tests"
    echo "  --cleanup   Cleanup test environment after running tests"
    echo "  --help      Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 --setup all --cleanup"
    echo "  $0 unit"
    echo "  $0 e2e"
}

# Parse command line arguments
SETUP_ENV=false
CLEANUP_ENV=false
TEST_TYPE="all"

while [[ $# -gt 0 ]]; do
    case $1 in
        --setup)
            SETUP_ENV=true
            shift
            ;;
        --cleanup)
            CLEANUP_ENV=true
            shift
            ;;
        --help)
            show_usage
            exit 0
            ;;
        unit|integration|e2e|performance|all)
            TEST_TYPE=$1
            shift
            ;;
        *)
            print_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Main execution
main() {
    print_status "Starting test execution..."
    print_status "Test type: $TEST_TYPE"
    
    # Setup test environment if requested
    if [ "$SETUP_ENV" = true ]; then
        setup_test_env || exit 1
    fi
    
    # Run tests based on type
    case $TEST_TYPE in
        unit)
            run_backend_tests && run_frontend_tests
            ;;
        integration)
            run_backend_tests
            ;;
        e2e)
            run_e2e_tests
            ;;
        performance)
            run_performance_tests
            ;;
        all)
            run_backend_tests && run_frontend_tests && run_e2e_tests
            ;;
        *)
            print_error "Invalid test type: $TEST_TYPE"
            show_usage
            exit 1
            ;;
    esac
    
    TEST_RESULT=$?
    
    # Cleanup test environment if requested
    if [ "$CLEANUP_ENV" = true ]; then
        cleanup_test_env
    fi
    
    if [ $TEST_RESULT -eq 0 ]; then
        print_success "All tests completed successfully!"
    else
        print_error "Some tests failed!"
        exit 1
    fi
}

# Run main function
main