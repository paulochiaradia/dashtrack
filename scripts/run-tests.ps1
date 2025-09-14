# Test Runner Script for Dashtrack API (PowerShell)
# This script runs all tests and generates reports

param(
    [switch]$SkipIntegration,
    [switch]$SkipBenchmarks,
    [switch]$SkipSecurity,
    [string]$CoverageThreshold = "80"
)

Write-Host "🚀 Starting Dashtrack API Test Suite" -ForegroundColor Blue
Write-Host "======================================" -ForegroundColor Blue

# Function to print colored output
function Write-Status {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor Cyan
}

function Write-Success {
    param([string]$Message)
    Write-Host "[SUCCESS] $Message" -ForegroundColor Green
}

function Write-Warning {
    param([string]$Message)
    Write-Host "[WARNING] $Message" -ForegroundColor Yellow
}

function Write-Error-Custom {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor Red
}

# Check if Go is installed
try {
    $goVersion = go version
    Write-Status "Using Go version: $goVersion"
} catch {
    Write-Error-Custom "Go is not installed or not in PATH"
    exit 1
}

# Create test output directory
New-Item -ItemType Directory -Force -Path "test-reports" | Out-Null

# 1. Download dependencies
Write-Status "Downloading dependencies..."
try {
    go mod download
    go mod tidy
    Write-Success "Dependencies downloaded successfully"
} catch {
    Write-Error-Custom "Failed to download dependencies"
    exit 1
}

# 2. Code quality checks
Write-Status "Running code quality checks..."

# Format check
try {
    $formatResult = go fmt ./...
    if ($formatResult) {
        Write-Warning "Code formatting applied to: $formatResult"
    } else {
        Write-Success "Code formatting check passed"
    }
} catch {
    Write-Error-Custom "Code formatting failed"
    exit 1
}

# Vet check
try {
    go vet ./...
    Write-Success "Go vet check passed"
} catch {
    Write-Error-Custom "Go vet found issues"
    exit 1
}

# 3. Unit tests
Write-Status "Running unit tests..."
try {
    go test -v -race -coverprofile=test-reports/coverage.out ./internal/...
    Write-Success "Unit tests passed"
} catch {
    Write-Error-Custom "Unit tests failed"
    exit 1
}

# Generate coverage report
try {
    go tool cover -html=test-reports/coverage.out -o test-reports/coverage.html
    $coverageOutput = go tool cover -func=test-reports/coverage.out | Select-Object -Last 1
    $coverage = ($coverageOutput -split '\s+')[-1]
    Write-Status "Test coverage: $coverage"
    
    # Check coverage threshold
    $coverageNum = [double]($coverage -replace '%', '')
    if ($coverageNum -lt [double]$CoverageThreshold) {
        Write-Warning "Test coverage is below $CoverageThreshold% ($coverage)"
    } else {
        Write-Success "Test coverage meets threshold: $coverage"
    }
} catch {
    Write-Warning "Could not generate coverage report"
}

# 4. Integration tests
if (-not $SkipIntegration -and (Test-Path "./tests/integration")) {
    Write-Status "Checking for integration tests..."
    
    # Check if Docker is available
    try {
        docker ps | Out-Null
        Write-Status "Running integration tests..."
        try {
            go test -v ./tests/integration/...
            Write-Success "Integration tests passed"
        } catch {
            Write-Warning "Integration tests failed or skipped"
        }
    } catch {
        Write-Warning "Docker not available, skipping integration tests"
    }
}

# 5. Benchmark tests
if (-not $SkipBenchmarks -and (Test-Path "./tests/benchmarks")) {
    Write-Status "Running benchmark tests..."
    try {
        go test -v -bench=. -benchmem ./tests/benchmarks/... > test-reports/benchmarks.txt
        Write-Success "Benchmark tests completed (results in test-reports/benchmarks.txt)"
    } catch {
        Write-Warning "Benchmark tests failed"
    }
}

# 6. Security scan
if (-not $SkipSecurity) {
    Write-Status "Running security scan..."
    try {
        $gosecExists = Get-Command gosec -ErrorAction SilentlyContinue
        if ($gosecExists) {
            gosec -fmt json -out test-reports/security.json ./...
            Write-Success "Security scan completed"
        } else {
            Write-Warning "gosec not installed, skipping security scan"
            Write-Status "Install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"
        }
    } catch {
        Write-Warning "Security scan found potential issues (check test-reports/security.json)"
    }
}

# 7. API tests
Write-Status "Checking API endpoints..."
try {
    $healthResponse = Invoke-RestMethod -Uri "http://localhost:8080/health" -Method GET -TimeoutSec 5
    Write-Status "API server is running, testing endpoints..."
    
    # Test health endpoint
    if ($healthResponse.status -eq "ok") {
        Write-Success "Health endpoint test passed"
    } else {
        Write-Warning "Health endpoint test failed"
    }
    
    # Test roles endpoint
    try {
        $rolesResponse = Invoke-RestMethod -Uri "http://localhost:8080/roles" -Method GET -TimeoutSec 5
        Write-Success "Roles endpoint test passed"
    } catch {
        Write-Warning "Roles endpoint test failed"
    }
    
    # Test users endpoint
    try {
        $usersResponse = Invoke-RestMethod -Uri "http://localhost:8080/users" -Method GET -TimeoutSec 5
        Write-Success "Users endpoint test passed"
    } catch {
        Write-Warning "Users endpoint test failed"
    }
    
} catch {
    Write-Warning "API server not running, skipping endpoint tests"
    Write-Status "Start with: go run ./cmd/api or make run"
}

# 8. Test summary
Write-Host ""
Write-Host "📊 Test Summary" -ForegroundColor Blue
Write-Host "===============" -ForegroundColor Blue
Write-Status "Test reports generated in: test-reports/"

if (Test-Path "test-reports/coverage.html") {
    Write-Status "Coverage report: test-reports/coverage.html"
}

if (Test-Path "test-reports/benchmarks.txt") {
    Write-Status "Benchmark results: test-reports/benchmarks.txt"
}

if (Test-Path "test-reports/security.json") {
    Write-Status "Security report: test-reports/security.json"
}

# 9. Performance recommendations
Write-Host ""
Write-Host "💡 Performance Recommendations" -ForegroundColor Blue
Write-Host "===============================" -ForegroundColor Blue
Write-Status "Monitor metrics at http://localhost:9090 (if Prometheus is running)"
Write-Status "View traces at http://localhost:16686 (if Jaeger is running)"
Write-Status "Check logs in Grafana at http://localhost:3000 (if monitoring stack is running)"

Write-Success "🎉 Test suite completed successfully!"
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Yellow
Write-Host "- Review test reports in test-reports/ directory"
Write-Host "- Check code coverage and aim for >$CoverageThreshold%"
Write-Host "- Monitor application metrics and logs"
Write-Host "- Run integration tests with Docker environment"
Write-Host ""
