#!/bin/bash

# Comprehensive test runner for Venue Discovery Platform
# This script sets up the test environment and runs all tests

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
TEST_DB_NAME="voting_app_test"
MAIN_DB_NAME="voting_app"
DB_USER="postgres"
DB_HOST="localhost"
DB_PORT="5432"

# Functions
print_header() {
    echo -e "\n${BLUE}============================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}============================================${NC}\n"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

# Check dependencies
check_dependencies() {
    print_header "Checking Dependencies"
    
    # Check Go
    if command -v go &> /dev/null; then
        GO_VERSION=$(go version | cut -d' ' -f3)
        print_success "Go found: $GO_VERSION"
    else
        print_error "Go not found. Please install Go 1.17 or later."
        exit 1
    fi
    
    # Check PostgreSQL
    if command -v psql &> /dev/null; then
        PG_VERSION=$(psql --version | cut -d' ' -f3)
        print_success "PostgreSQL found: $PG_VERSION"
    else
        print_error "PostgreSQL not found. Please install PostgreSQL."
        exit 1
    fi
    
    # Check if PostgreSQL is running
    if pg_isready -h $DB_HOST -p $DB_PORT &> /dev/null; then
        print_success "PostgreSQL is running"
    else
        print_error "PostgreSQL is not running. Please start PostgreSQL service."
        exit 1
    fi
    
    # Check test framework dependencies
    echo "Checking Go test dependencies..."
    go mod download
    go mod tidy
    print_success "Go dependencies updated"
}

# Setup test database
setup_test_database() {
    print_header "Setting Up Test Database"
    
    # Drop existing test database if it exists
    echo "Dropping existing test database (if exists)..."
    psql -h $DB_HOST -U $DB_USER -c "DROP DATABASE IF EXISTS $TEST_DB_NAME;" postgres &> /dev/null || true
    
    # Create test database
    echo "Creating test database..."
    psql -h $DB_HOST -U $DB_USER -c "CREATE DATABASE $TEST_DB_NAME;" postgres
    print_success "Test database '$TEST_DB_NAME' created"
    
    # Check if enhanced schema exists
    if [ -f "enhanced_schema.sql" ]; then
        echo "Running enhanced schema migrations..."
        psql -h $DB_HOST -U $DB_USER -d $TEST_DB_NAME -f enhanced_schema.sql
        print_success "Enhanced schema applied"
    else
        print_warning "enhanced_schema.sql not found, using basic schema"
    fi
    
    # Apply PostGIS extension if available (for geospatial features)
    echo "Adding PostGIS extension (if available)..."
    psql -h $DB_HOST -U $DB_USER -d $TEST_DB_NAME -c "CREATE EXTENSION IF NOT EXISTS postgis;" &> /dev/null || print_warning "PostGIS extension not available"
}

# Run tests with different configurations
run_unit_tests() {
    print_header "Running Unit Tests"
    
    echo "Running unit tests for app modules..."
    if go test -v -short ./app/...; then
        print_success "Unit tests passed"
        return 0
    else
        print_error "Unit tests failed"
        return 1
    fi
}

run_integration_tests() {
    print_header "Running Integration Tests"
    
    # Set test environment
    export GIN_MODE=test
    export DB_NAME=$TEST_DB_NAME
    
    echo "Running integration tests..."
    if go test -v -run "TestSuite" ./tests/ -timeout 10m; then
        print_success "Integration tests passed"
        return 0
    else
        print_error "Integration tests failed"
        return 1
    fi
}

run_e2e_tests() {
    print_header "Running End-to-End Tests"
    
    # Set test environment
    export GIN_MODE=test
    export DB_NAME=$TEST_DB_NAME
    
    echo "Running full end-to-end test suite..."
    if go test -v ./tests/ -timeout 30m; then
        print_success "End-to-end tests passed"
        return 0
    else
        print_error "End-to-end tests failed"
        return 1
    fi
}

run_performance_tests() {
    print_header "Running Performance Tests"
    
    echo "Running benchmark tests..."
    if go test -bench=. -benchmem ./tests/ -timeout 15m; then
        print_success "Performance tests completed"
        return 0
    else
        print_error "Performance tests failed"
        return 1
    fi
}

generate_coverage_report() {
    print_header "Generating Coverage Report"
    
    mkdir -p coverage
    
    echo "Running tests with coverage..."
    go test -v -coverprofile=coverage/coverage.out ./app/... ./tests/...
    
    if [ -f "coverage/coverage.out" ]; then
        # Generate HTML report
        go tool cover -html=coverage/coverage.out -o coverage/coverage.html
        
        # Show coverage summary
        go tool cover -func=coverage/coverage.out | tail -1
        
        print_success "Coverage report generated: coverage/coverage.html"
        
        # Open coverage report if on macOS
        if [[ "$OSTYPE" == "darwin"* ]]; then
            echo "Opening coverage report..."
            open coverage/coverage.html
        fi
    else
        print_error "Coverage file not generated"
    fi
}

# Test specific features
test_venue_discovery() {
    print_header "Testing Venue Discovery Features"
    go test -v -run "TestVenueDiscovery" ./tests/ -timeout 5m
}

test_review_system() {
    print_header "Testing Review System"
    go test -v -run "TestReviewSystem" ./tests/ -timeout 5m
}

test_voting_system() {
    print_header "Testing Voting System"
    go test -v -run "TestLegacyVotingSystem|TestEnhancedVotingCampaigns" ./tests/ -timeout 5m
}

test_analytics() {
    print_header "Testing Analytics System"
    go test -v -run "TestAnalyticsSystem" ./tests/ -timeout 5m
}

test_recommendations() {
    print_header "Testing Recommendation Engine"
    go test -v -run "TestRecommendationSystem" ./tests/ -timeout 5m
}

# Cleanup function
cleanup() {
    print_header "Cleaning Up"
    
    if [ "$KEEP_TEST_DB" != "true" ]; then
        echo "Dropping test database..."
        psql -h $DB_HOST -U $DB_USER -c "DROP DATABASE IF EXISTS $TEST_DB_NAME;" postgres &> /dev/null || true
        print_success "Test database cleaned up"
    else
        print_warning "Test database preserved (KEEP_TEST_DB=true)"
    fi
    
    # Clean test artifacts
    rm -rf tmp/
    echo "Temporary files cleaned"
}

# Show usage
show_usage() {
    echo "Usage: $0 [OPTIONS] [TEST_TYPE]"
    echo ""
    echo "Test Types:"
    echo "  all              Run all tests (default)"
    echo "  unit             Run unit tests only"
    echo "  integration      Run integration tests only"
    echo "  e2e              Run end-to-end tests only"
    echo "  performance      Run performance/benchmark tests"
    echo "  venue            Test venue discovery features"
    echo "  review           Test review system"
    echo "  voting           Test voting system"
    echo "  analytics        Test analytics system"
    echo "  recommendations  Test recommendation engine"
    echo "  coverage         Generate coverage report"
    echo ""
    echo "Options:"
    echo "  -h, --help       Show this help message"
    echo "  --keep-db        Keep test database after tests"
    echo "  --no-setup       Skip database setup"
    echo "  --verbose        Verbose output"
    echo "  --parallel       Run tests in parallel"
    echo ""
    echo "Environment Variables:"
    echo "  DB_HOST          Database host (default: localhost)"
    echo "  DB_PORT          Database port (default: 5432)"
    echo "  DB_USER          Database user (default: postgres)"
    echo "  KEEP_TEST_DB     Keep test database (default: false)"
    echo "  TEST_TIMEOUT     Test timeout (default: 30m)"
    echo ""
    echo "Examples:"
    echo "  $0                    # Run all tests"
    echo "  $0 unit              # Run unit tests only"
    echo "  $0 --keep-db e2e     # Run e2e tests and keep database"
    echo "  DB_USER=myuser $0    # Run with custom database user"
}

# Parse command line arguments
SKIP_SETUP=false
VERBOSE=false
PARALLEL=false
TEST_TYPE="all"

while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_usage
            exit 0
            ;;
        --keep-db)
            export KEEP_TEST_DB=true
            shift
            ;;
        --no-setup)
            SKIP_SETUP=true
            shift
            ;;
        --verbose)
            VERBOSE=true
            shift
            ;;
        --parallel)
            PARALLEL=true
            shift
            ;;
        unit|integration|e2e|performance|venue|review|voting|analytics|recommendations|coverage|all)
            TEST_TYPE="$1"
            shift
            ;;
        *)
            echo "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Set verbose mode
if [ "$VERBOSE" = true ]; then
    set -x
fi

# Main execution
main() {
    print_header "Venue Discovery Platform - Test Runner"
    
    # Check environment
    check_dependencies
    
    # Setup test database unless skipped
    if [ "$SKIP_SETUP" = false ]; then
        setup_test_database
    fi
    
    # Set up trap for cleanup
    trap cleanup EXIT
    
    # Track test results
    TESTS_PASSED=0
    TESTS_FAILED=0
    
    # Run specified tests
    case $TEST_TYPE in
        "unit")
            if run_unit_tests; then ((TESTS_PASSED++)); else ((TESTS_FAILED++)); fi
            ;;
        "integration")
            if run_integration_tests; then ((TESTS_PASSED++)); else ((TESTS_FAILED++)); fi
            ;;
        "e2e")
            if run_e2e_tests; then ((TESTS_PASSED++)); else ((TESTS_FAILED++)); fi
            ;;
        "performance")
            if run_performance_tests; then ((TESTS_PASSED++)); else ((TESTS_FAILED++)); fi
            ;;
        "venue")
            if test_venue_discovery; then ((TESTS_PASSED++)); else ((TESTS_FAILED++)); fi
            ;;
        "review")
            if test_review_system; then ((TESTS_PASSED++)); else ((TESTS_FAILED++)); fi
            ;;
        "voting")
            if test_voting_system; then ((TESTS_PASSED++)); else ((TESTS_FAILED++)); fi
            ;;
        "analytics")
            if test_analytics; then ((TESTS_PASSED++)); else ((TESTS_FAILED++)); fi
            ;;
        "recommendations")
            if test_recommendations; then ((TESTS_PASSED++)); else ((TESTS_FAILED++)); fi
            ;;
        "coverage")
            generate_coverage_report
            ;;
        "all")
            if run_unit_tests; then ((TESTS_PASSED++)); else ((TESTS_FAILED++)); fi
            if run_integration_tests; then ((TESTS_PASSED++)); else ((TESTS_FAILED++)); fi
            if run_e2e_tests; then ((TESTS_PASSED++)); else ((TESTS_FAILED++)); fi
            generate_coverage_report
            ;;
        *)
            print_error "Unknown test type: $TEST_TYPE"
            show_usage
            exit 1
            ;;
    esac
    
    # Show final results
    print_header "Test Results Summary"
    
    if [ $TESTS_FAILED -eq 0 ]; then
        print_success "All tests passed! ($TESTS_PASSED passed)"
        exit 0
    else
        print_error "$TESTS_FAILED test suite(s) failed, $TESTS_PASSED passed"
        exit 1
    fi
}

# Run main function
main "$@"
