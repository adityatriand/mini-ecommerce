#!/bin/bash

# Pre-commit hook for Mini E-Commerce Microservices
# This script runs tests and linting for modified services

set -e

echo "🔍 Running pre-commit checks for Mini E-Commerce Microservices..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

# Get list of modified files
MODIFIED_FILES=$(git diff --cached --name-only --diff-filter=ACM)

if [ -z "$MODIFIED_FILES" ]; then
    print_warning "No modified files to check"
    exit 0
fi

print_status "Modified files:"
echo "$MODIFIED_FILES"

# Determine which services are affected
AFFECTED_SERVICES=()

for file in $MODIFIED_FILES; do
    if [[ $file == services/gateway/* ]]; then
        if [[ ! " ${AFFECTED_SERVICES[@]} " =~ " gateway " ]]; then
            AFFECTED_SERVICES+=("gateway")
        fi
    elif [[ $file == services/user/* ]]; then
        if [[ ! " ${AFFECTED_SERVICES[@]} " =~ " user " ]]; then
            AFFECTED_SERVICES+=("user")
        fi
    elif [[ $file == services/product/* ]]; then
        if [[ ! " ${AFFECTED_SERVICES[@]} " =~ " product " ]]; then
            AFFECTED_SERVICES+=("product")
        fi
    elif [[ $file == services/order/* ]]; then
        if [[ ! " ${AFFECTED_SERVICES[@]} " =~ " order " ]]; then
            AFFECTED_SERVICES+=("order")
        fi
    elif [[ $file == shared/* ]]; then
        # If shared libraries are modified, test all services
        AFFECTED_SERVICES=("gateway" "user" "product" "order")
        break
    fi
done

if [ ${#AFFECTED_SERVICES[@]} -eq 0 ]; then
    print_success "No service files modified, skipping tests"
    exit 0
fi

print_status "Affected services: ${AFFECTED_SERVICES[*]}"

# Check if Go is available
if ! command -v go &> /dev/null; then
    print_error "Go is not installed or not in PATH"
    exit 1
fi

# Check if we're in the right directory
if [ ! -f "go.mod" ]; then
    print_error "go.mod not found. Please run this script from the project root"
    exit 1
fi

# Function to run tests for a specific service
run_service_tests() {
    local service=$1
    print_status "Running tests for $service service..."
    
    # Change to service directory
    cd "services/$service"
    
    # Run go mod tidy
    print_status "Running go mod tidy for $service..."
    go mod tidy
    
    # Run go vet
    print_status "Running go vet for $service..."
    if ! go vet ./...; then
        print_error "go vet failed for $service service"
        cd ../..
        return 1
    fi
    
    # Run go fmt check
    print_status "Checking go fmt for $service..."
    if [ -n "$(gofmt -l .)" ]; then
        print_error "Code is not formatted. Please run 'go fmt ./...' for $service service"
        cd ../..
        return 1
    fi
    
    # Run tests
    print_status "Running tests for $service..."
    if ! go test -v ./...; then
        print_error "Tests failed for $service service"
        cd ../..
        return 1
    fi
    
    # Run test coverage
    print_status "Running test coverage for $service..."
    if ! go test -coverprofile=coverage.out ./...; then
        print_error "Test coverage failed for $service service"
        cd ../..
        return 1
    fi
    
    # Show coverage
    coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
    print_success "$service service tests passed with $coverage coverage"
    
    # Clean up
    rm -f coverage.out
    cd ../..
}

# Function to run shared library tests
run_shared_tests() {
    print_status "Running tests for shared libraries..."
    
    cd shared
    
    # Run go mod tidy
    go mod tidy
    
    # Run go vet
    if ! go vet ./...; then
        print_error "go vet failed for shared libraries"
        cd ..
        return 1
    fi
    
    # Run go fmt check
    if [ -n "$(gofmt -l .)" ]; then
        print_error "Code is not formatted. Please run 'go fmt ./...' for shared libraries"
        cd ..
        return 1
    fi
    
    # Run tests
    if ! go test -v ./...; then
        print_error "Tests failed for shared libraries"
        cd ..
        return 1
    fi
    
    print_success "Shared libraries tests passed"
    cd ..
}

# Run tests for affected services
FAILED_SERVICES=()

for service in "${AFFECTED_SERVICES[@]}"; do
    if ! run_service_tests "$service"; then
        FAILED_SERVICES+=("$service")
    fi
done

# Run shared library tests if shared files were modified
if [[ " ${MODIFIED_FILES[@]} " =~ " shared/" ]]; then
    if ! run_shared_tests; then
        print_error "Shared library tests failed"
        exit 1
    fi
fi

# Check if any services failed
if [ ${#FAILED_SERVICES[@]} -gt 0 ]; then
    print_error "The following services failed tests: ${FAILED_SERVICES[*]}"
    print_error "Please fix the issues before committing"
    exit 1
fi

# Run integration tests if available
if [ -f "Makefile" ]; then
    print_status "Running integration tests..."
    if make test-integration 2>/dev/null; then
        print_success "Integration tests passed"
    else
        print_warning "Integration tests not available or failed (this is optional)"
    fi
fi

# Check for TODO/FIXME comments in staged files
print_status "Checking for TODO/FIXME comments..."
TODO_FOUND=false
for file in $MODIFIED_FILES; do
    if [ -f "$file" ]; then
        if grep -n "TODO\|FIXME" "$file" > /dev/null 2>&1; then
            print_warning "TODO/FIXME found in $file:"
            grep -n "TODO\|FIXME" "$file"
            TODO_FOUND=true
        fi
    fi
done

if [ "$TODO_FOUND" = true ]; then
    print_warning "TODO/FIXME comments found. Consider addressing them before committing."
fi

# Final success message
print_success "All pre-commit checks passed! ✅"
print_success "Services tested: ${AFFECTED_SERVICES[*]}"

echo ""
echo "🚀 Ready to commit!"
echo ""


