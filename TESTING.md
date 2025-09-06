# Testing Documentation - Venue Discovery Platform

## Overview

This document provides comprehensive information about testing the enhanced venue discovery and rating platform. The testing suite covers all features from basic venue search to advanced analytics and recommendation systems.

## Test Structure

### Test Organization

```
tests/
├── setup_test.go           # Test suite setup and infrastructure
├── venue_e2e_test.go       # Venue discovery and management tests
├── review_e2e_test.go      # Review and rating system tests
├── voting_e2e_test.go      # Legacy voting and campaign tests
├── analytics_e2e_test.go   # Analytics and recommendation tests
└── integration_e2e_test.go # Cross-feature integration tests
```

### Test Categories

1. **Unit Tests** - Individual function/method testing
2. **Integration Tests** - Module interaction testing
3. **End-to-End Tests** - Complete user journey testing
4. **Performance Tests** - Load and benchmark testing

## Test Features Covered

### ✅ Venue Discovery System
- **Venue Search** - Text search, filters, pagination, sorting
- **Location-based Discovery** - Nearby venues, radius search, distance calculation
- **Category Filtering** - Restaurant types, price ranges, amenities
- **Featured Venues** - Curated venue promotion
- **Venue Details** - Complete venue information retrieval
- **Venue Management** - CRUD operations for venues

### ✅ Review & Rating System
- **Review Creation** - Multi-dimensional ratings, photos, visit context
- **Review Validation** - Rating ranges, duplicate prevention, moderation
- **Review Discovery** - Filtering, sorting, pagination
- **Review Analytics** - Summaries, distributions, trending
- **Review Helpfulness** - Community voting on review quality
- **User Review History** - Personal review management

### ✅ Enhanced Voting System
- **Legacy Compatibility** - Original talent voting system
- **Voting Campaigns** - "Best of" competitions for venues
- **Campaign Management** - Time-bound voting, rules, results
- **Vote Validation** - Duplicate prevention, eligibility checks
- **Results Calculation** - Winner determination, statistics

### ✅ Analytics & Intelligence
- **Venue Analytics** - Performance metrics, engagement tracking
- **Search Analytics** - Query trends, click-through rates
- **User Behavior** - Activity patterns, preferences
- **Platform Metrics** - Overall system health and usage
- **Performance Tracking** - Response times, throughput

### ✅ Recommendation Engine
- **Personalized Recommendations** - AI-driven venue suggestions
- **Similar Venue Discovery** - Content-based filtering
- **Social Recommendations** - Friend-based suggestions
- **Context-aware Scoring** - Time, location, occasion factors
- **Preference Learning** - User behavior analysis

### ✅ Geolocation Services
- **Distance Calculations** - Precise geographical measurements
- **Boundary Operations** - Geographical bounds, area searches
- **Optimal Meeting Points** - Group location optimization
- **Location Validation** - Coordinate verification, geocoding

## Test Environment Setup

### Prerequisites

1. **Go 1.17+** - Programming language
2. **PostgreSQL 12+** - Database system
3. **PostGIS Extension** - Geospatial operations (optional)
4. **Git** - Version control

### Database Setup

```bash
# Create test database
createdb voting_app_test

# Apply enhanced schema
psql -d voting_app_test -f enhanced_schema.sql

# Add PostGIS extension (optional)
psql -d voting_app_test -c "CREATE EXTENSION IF NOT EXISTS postgis;"
```

### Environment Configuration

Create `test.env` file:

```env
# Test Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=voting_app_test
DB_USER=postgres
DB_PASS=your_password

# Test Configuration
GIN_MODE=test
LOG_LEVEL=error
TEST_DB_RESET=true
```

## Running Tests

### Quick Start

```bash
# Run all tests
./run_tests.sh

# Run specific test category
./run_tests.sh unit
./run_tests.sh integration
./run_tests.sh e2e

# Run with coverage report
./run_tests.sh coverage
```

### Using Make Commands

```bash
# Setup and run all tests
make test-all

# Individual test types
make test-unit
make test-integration
make test-e2e

# Generate coverage report
make test-coverage

# Watch mode for development
make test-watch
```

### Manual Test Execution

```bash
# Run venue discovery tests
go test -v -run "TestVenueDiscovery" ./tests/

# Run review system tests
go test -v -run "TestReviewSystem" ./tests/

# Run analytics tests
go test -v -run "TestAnalyticsSystem" ./tests/

# Run with race detection
go test -race ./tests/

# Run with benchmarks
go test -bench=. ./tests/
```

## Test Configuration Options

### Command Line Flags

```bash
# Keep test database after tests
./run_tests.sh --keep-db e2e

# Skip database setup
./run_tests.sh --no-setup unit

# Verbose output
./run_tests.sh --verbose all

# Run specific feature tests
./run_tests.sh venue      # Venue discovery only
./run_tests.sh review     # Review system only
./run_tests.sh voting     # Voting system only
./run_tests.sh analytics  # Analytics only
```

### Environment Variables

```bash
# Custom database configuration
DB_HOST=testdb.example.com ./run_tests.sh

# Keep test database for debugging
KEEP_TEST_DB=true ./run_tests.sh

# Custom test timeout
TEST_TIMEOUT=45m ./run_tests.sh performance
```

## Test Data & Fixtures

### Automatic Test Data

The test suite automatically creates:

- **2 Test Users** - `test_user_1`, `test_user_2`
- **1 Test City** - San Francisco with coordinates
- **1 Test Category** - Restaurant category
- **2 Test Venues** - Sample restaurants with different ratings
- **Sample Reviews** - Multi-dimensional ratings and text
- **Analytics Data** - Performance metrics, search logs

### Custom Test Data

You can add custom test data in `setupTestData()`:

```go
// Add custom venue
_, err := suite.db.Exec(`INSERT INTO venues 
    (name, address, city_id, latitude, longitude, category_id) 
    VALUES ($1, $2, $3, $4, $5, $6)`, 
    "Custom Restaurant", "123 Test St", 1, 37.7749, -122.4194, 1)
```

## Test Scenarios Covered

### User Journey Tests

1. **Discovery Journey**
   - Search for restaurants → View details → Check reviews → Compare nearby options

2. **Review Journey**
   - Visit venue → Write detailed review → See review in listings → Others vote helpful

3. **Collection Journey**
   - Create favorite list → Add venues → Share publicly → Browse others' collections

4. **Voting Journey**
   - Participate in talent voting → Vote in venue campaigns → View results

### Edge Cases & Error Handling

- **Invalid Input Validation** - Malformed data, out-of-range values
- **Authentication Errors** - Missing tokens, expired sessions
- **Database Constraints** - Duplicate data, foreign key violations
- **Geographic Edge Cases** - Invalid coordinates, extreme distances
- **Concurrent Operations** - Race conditions, data consistency

### Performance Testing

- **Load Testing** - High concurrent user simulation
- **Stress Testing** - System breaking point identification
- **Benchmark Testing** - Response time measurements
- **Memory Testing** - Resource usage optimization

## Continuous Integration

### GitHub Actions (Example)

```yaml
name: Test Suite
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgis/postgis:13-3.1
        env:
          POSTGRES_PASSWORD: postgres
        options: --health-cmd pg_isready --health-interval 10s
    
    steps:
    - uses: actions/checkout@v2
    - uses: actions/setup-go@v2
      with:
        go-version: 1.17
    
    - name: Run tests
      run: ./run_tests.sh all
      env:
        DB_HOST: localhost
        DB_USER: postgres
        DB_PASS: postgres
```

## Debugging Tests

### Test Database Inspection

```bash
# Keep database after test failure
KEEP_TEST_DB=true ./run_tests.sh

# Connect to test database
psql -d voting_app_test

# Inspect test data
SELECT * FROM venues;
SELECT * FROM venue_reviews;
SELECT * FROM venue_analytics;
```

### Verbose Test Output

```bash
# Enable verbose Go test output
go test -v ./tests/

# Enable SQL query logging
export LOG_LEVEL=debug
./run_tests.sh
```

### Coverage Analysis

```bash
# Generate detailed coverage
make test-coverage

# View coverage in browser
open coverage/coverage.html

# Show uncovered functions
go tool cover -func=coverage/coverage.out | grep -v "100.0%"
```

## Test Maintenance

### Adding New Tests

1. **Create test function** in appropriate `*_test.go` file
2. **Follow naming convention** - `Test[Feature][Scenario]`
3. **Add to test suite** if it requires database setup
4. **Update this documentation** with new test coverage

### Mock Services

For external service testing:

```go
// Mock recommendation service
type MockRecommendationEngine struct{}

func (m *MockRecommendationEngine) GetPersonalizedRecommendations(ctx RecommendationContext) ([]RecommendationScore, error) {
    // Return test data
    return []RecommendationScore{}, nil
}
```

### Test Data Cleanup

```go
func (suite *TestSuite) TearDownTest() {
    // Clean up after each test
    suite.cleanupTestData()
}
```

## Performance Benchmarks

### Target Performance Metrics

- **Venue Search** - < 100ms for 10,000 venues
- **Nearby Search** - < 50ms within 10km radius
- **Review Creation** - < 200ms with validation
- **Recommendation Generation** - < 500ms for personalized results
- **Analytics Queries** - < 1s for daily summits

### Benchmark Tests

```bash
# Run performance benchmarks
go test -bench=BenchmarkVenueSearch ./tests/
go test -bench=BenchmarkRecommendation ./tests/
go test -bench=. ./tests/ # All benchmarks
```

## Troubleshooting

### Common Issues

1. **Database Connection** - Check PostgreSQL service, credentials
2. **Missing Dependencies** - Run `go mod download`
3. **Port Conflicts** - Ensure test ports are available
4. **Permission Errors** - Check database user permissions
5. **Timeout Issues** - Increase test timeout for slow systems

### Test Isolation

Tests are designed to be independent:
- Each test starts with clean database state
- Test data is cleaned up after each test
- No shared state between test functions

### Memory Leaks

Monitor memory usage during tests:

```bash
# Run with memory profiling
go test -memprofile=mem.prof ./tests/

# Analyze memory usage
go tool pprof mem.prof
```

## Future Enhancements

### Planned Test Additions

1. **Mobile API Testing** - iOS/Android specific endpoints
2. **Real-time Features** - WebSocket connection testing
3. **Caching Layer** - Redis integration testing
4. **External API Mocks** - Google Maps, payment processors
5. **Security Testing** - SQL injection, XSS prevention
6. **Accessibility Testing** - API response format validation

### Test Infrastructure Improvements

1. **Parallel Test Execution** - Faster test suite completion
2. **Docker Test Environment** - Consistent testing across environments
3. **Property-based Testing** - Automated edge case discovery
4. **Mutation Testing** - Test quality assessment
5. **Contract Testing** - API compatibility verification

---

This comprehensive testing suite ensures the venue discovery platform is reliable, performant, and ready for production deployment. The tests cover all major features and edge cases, providing confidence in system stability and user experience quality.
