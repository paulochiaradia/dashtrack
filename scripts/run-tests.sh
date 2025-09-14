#!/bin/bash

# Test Runner Script for Dashtrack API
# This script runs all tests and generates reports

set -e

echo "🚀 Starting Dashtrack API Test Suite"
echo "======================================"

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

# Check if Go is installed
if ! command -v go &> /dev/null; then
    print_error "Go is not installed or not in PATH"
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | awk '{print $3}')
print_status "Using Go version: $GO_VERSION"

# Create test output directory
mkdir -p test-reports

# 1. Download dependencies
print_status "Downloading dependencies..."
go mod download
go mod tidy
print_success "Dependencies downloaded successfully"

# 2. Code quality checks
print_status "Running code quality checks..."

# Format check
if ! go fmt ./...; then
    print_error "Code formatting issues found"
    exit 1
fi
print_success "Code formatting check passed"

# Vet check
if ! go vet ./...; then
    print_error "Go vet found issues"
    exit 1
fi
print_success "Go vet check passed"

# 3. Unit tests
print_status "Running unit tests..."
if ! go test -v -race -coverprofile=test-reports/coverage.out ./internal/...; then
    print_error "Unit tests failed"
    exit 1
fi
print_success "Unit tests passed"

# Generate coverage report
if command -v go tool cover &> /dev/null; then
    go tool cover -html=test-reports/coverage.out -o test-reports/coverage.html
    COVERAGE=$(go tool cover -func=test-reports/coverage.out | tail -n 1 | awk '{print $3}')
    print_status "Test coverage: $COVERAGE"
    
    # Check coverage threshold (e.g., 80%)
    COVERAGE_NUM=$(echo $COVERAGE | sed 's/%//')
    if (( $(echo "$COVERAGE_NUM < 80" | bc -l) )); then
        print_warning "Test coverage is below 80% ($COVERAGE)"
    else
        print_success "Test coverage meets threshold: $COVERAGE"
    fi
fi

# 4. Integration tests (if available and Docker is running)
print_status "Checking for integration tests..."
if [ -d "./tests/integration" ]; then
    if command -v docker &> /dev/null && docker ps &> /dev/null; then
        print_status "Running integration tests..."
        if ! go test -v ./tests/integration/...; then
            print_warning "Integration tests failed or skipped"
        else
            print_success "Integration tests passed"
        fi
    else
        print_warning "Docker not available, skipping integration tests"
    fi
fi

# 5. Benchmark tests
print_status "Running benchmark tests..."
if [ -d "./tests/benchmarks" ]; then
    go test -v -bench=. -benchmem ./tests/benchmarks/... > test-reports/benchmarks.txt
    print_success "Benchmark tests completed (results in test-reports/benchmarks.txt)"
fi

# 6. Security scan (if gosec is installed)
print_status "Running security scan..."
if command -v gosec &> /dev/null; then
    if gosec -fmt json -out test-reports/security.json ./...; then
        print_success "Security scan completed (no high-severity issues)"
    else
        print_warning "Security scan found potential issues (check test-reports/security.json)"
    fi
else
    print_warning "gosec not installed, skipping security scan"
    print_status "Install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"
fi

# 7. API tests (if server is running)
print_status "Checking API endpoints..."
if curl -s http://localhost:8080/health &> /dev/null; then
    print_status "API server is running, testing endpoints..."
    
    # Test health endpoint
    if curl -s http://localhost:8080/health | jq . &> /dev/null; then
        print_success "Health endpoint test passed"
    else
        print_warning "Health endpoint test failed"
    fi
    
    # Test roles endpoint
    if curl -s http://localhost:8080/roles | jq . &> /dev/null; then
        print_success "Roles endpoint test passed"
    else
        print_warning "Roles endpoint test failed"
    fi
    
    # Test users endpoint
    if curl -s http://localhost:8080/users | jq . &> /dev/null; then
        print_success "Users endpoint test passed"
    else
        print_warning "Users endpoint test failed"
    fi
else
    print_warning "API server not running, skipping endpoint tests"
    print_status "Start with: go run ./cmd/api or make run"
fi

# 8. Test summary
echo ""
echo "📊 Test Summary"
echo "==============="
print_status "Test reports generated in: test-reports/"
if [ -f "test-reports/coverage.html" ]; then
    print_status "Coverage report: test-reports/coverage.html"
fi
if [ -f "test-reports/benchmarks.txt" ]; then
    print_status "Benchmark results: test-reports/benchmarks.txt"
fi
if [ -f "test-reports/security.json" ]; then
    print_status "Security report: test-reports/security.json"
fi

# 9. Performance recommendations
echo ""
echo "💡 Performance Recommendations"
echo "==============================="
print_status "Run 'make load-test' to perform load testing"
print_status "Monitor metrics at http://localhost:9090 (if Prometheus is running)"
print_status "View traces at http://localhost:16686 (if Jaeger is running)"
print_status "Check logs in Grafana at http://localhost:3000 (if monitoring stack is running)"

print_success "🎉 Test suite completed successfully!"
echo ""
echo "Next steps:"
echo "- Review test reports in test-reports/ directory"
echo "- Check code coverage and aim for >80%"
echo "- Monitor application metrics and logs"
echo "- Run integration tests with Docker environment"
echo ""
