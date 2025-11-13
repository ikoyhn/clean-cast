# Test Suite Documentation

This document describes the comprehensive test suite created for the Clean Cast project.

## Test Coverage

### 1. HTTP Controller Tests (`internal/app/controller_test.go`)

Tests for HTTP endpoints and request handling:

- **Query Parameter Validation**: Tests for limit, date, and invalid parameters
- **Handler URL Generation**: Tests for HTTP/HTTPS scheme detection
- **Health Check Endpoints**: Validates root endpoint responses
- **Benchmarks**: Performance tests for query parameter validation

**Key Features**:
- Table-driven tests for comprehensive coverage
- HTTP test server simulation
- Edge case testing (invalid inputs, conflicting parameters)

### 2. Download Service Tests (`internal/services/downloader/ytdlp_test.go`)

Tests for YouTube video download logic:

- **File Existence Checks**: Verifies caching behavior
- **Concurrent Download Handling**: Tests mutex management
- **Callback System**: Tests download start/complete callbacks
- **File Extension Handling**: Tests .m4a extension stripping

**Key Features**:
- Concurrent access testing with goroutines
- Temporary directory management
- Mutex synchronization validation

### 3. YouTube API Tests (`internal/services/youtube/youtube_test.go`)

Tests for YouTube API integration (with mocking):

- **Channel Data Retrieval**: Tests caching and API interaction
- **Video Filtering**: Tests duration-based filtering
- **Duration Parsing**: Validates ISO 8601 duration parsing
- **Service Setup**: Tests YouTube API service initialization

**Key Features**:
- Mock-friendly design (skips tests if no API key)
- Duration comparison logic testing
- Cache hit behavior validation

### 4. Database Repository Tests (`internal/database/repository_test.go`)

Tests for database operations using in-memory SQLite:

- **Podcast CRUD Operations**: Create, Read, Update, Delete
- **Episode Management**: Episode existence, latest/oldest retrieval
- **Playback History**: Tracking and cleanup
- **Cleanup Jobs**: Tests cron-based file deletion

**Key Features**:
- In-memory SQLite for fast, isolated tests
- Fresh database for each test
- Complex query testing (filtering by duration, date)

### 5. Cache Tests (`internal/cache/cache_test.go`)

Tests for RSS feed caching system:

- **Get/Set Operations**: Basic cache operations
- **TTL Expiration**: Time-based cache invalidation
- **Concurrent Access**: Thread-safe operation validation
- **Cache Key Generation**: Consistent hashing tests

**Key Features**:
- Singleton pattern validation
- Concurrent read/write stress testing
- Cache hit/miss performance benchmarks

### 6. RSS Benchmark Tests (`internal/services/rss/rss_bench_test.go`)

Performance benchmarks for RSS feed generation:

- **Feed Size Variations**: Tests with 10, 50, 100, 500 episodes
- **Episode Filtering**: Benchmark filtering logic
- **Feed Type Comparison**: Channel vs Playlist performance

**Key Features**:
- Realistic test data generation
- Mixed content scenarios (short videos, private videos)
- Memory allocation tracking

## Running Tests

### Run All Tests
```bash
go test ./... -v
```

### Run Tests with Coverage
```bash
go test ./... -coverprofile=coverage.txt -covermode=atomic
go tool cover -html=coverage.txt -o coverage.html
```

### Run Specific Package
```bash
go test ./internal/cache -v
go test ./internal/database -v
```

### Run Benchmarks
```bash
# All benchmarks
go test -bench=. -benchmem ./...

# Specific package benchmarks
go test -bench=. -benchmem ./internal/cache
go test -bench=BenchmarkCache_Get_Hit -benchmem ./internal/cache
```

### Run with Race Detection
```bash
go test -race ./...
```

## CI/CD Integration

The test suite integrates with GitHub Actions (`.github/workflows/test.yml`):

- **Test Matrix**: Tests on Go 1.23.1 and 1.23.x
- **Code Coverage**: Automatically uploads to Codecov
- **Linting**: Uses golangci-lint for code quality
- **Security Scanning**: Runs gosec security scanner
- **Build Verification**: Multi-platform build checks (Linux, macOS, Windows)
- **Docker Build**: Tests Docker multi-arch builds

## Test Best Practices

### Table-Driven Tests
All tests use table-driven patterns for maintainability:

```go
tests := []struct {
    name     string
    input    string
    expected bool
}{
    {"valid input", "test", true},
    {"invalid input", "", false},
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        result := function(tt.input)
        assert.Equal(t, tt.expected, result)
    })
}
```

### In-Memory Testing
Database tests use in-memory SQLite for speed and isolation:

```go
testDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
```

### Concurrent Testing
Tests validate thread-safety with goroutines:

```go
b.RunParallel(func(pb *testing.PB) {
    for pb.Next() {
        cache.Get(feedType, id, params)
    }
})
```

## Test Infrastructure

### Dependencies
- **testify**: Assertions and test utilities
- **sqlite (in-memory)**: Fast database testing
- **httptest**: HTTP endpoint testing

### Mock Strategy
- YouTube API: Tests skip if no API key provided
- File System: Uses temporary directories
- Database: In-memory SQLite instances

## Benchmark Results

Benchmarks measure:
- **Operations per second**: Higher is better
- **Memory allocations**: Lower is better
- **Bytes allocated**: Lower is better

Example output:
```
BenchmarkCache_Get_Hit-8         10000000    120 ns/op    0 B/op    0 allocs/op
BenchmarkCache_Get_Miss-8         5000000    250 ns/op   64 B/op    2 allocs/op
```

## Code Quality Tools

### Linting
```bash
golangci-lint run --timeout=5m
```

Configuration in `.golangci.yml` enables:
- Error checking
- Security scanning
- Code simplification
- Style consistency

### Security Scanning
```bash
gosec ./...
```

## Continuous Improvement

To add new tests:

1. Create `*_test.go` file in the same package
2. Follow table-driven test pattern
3. Add benchmarks for performance-critical code
4. Use testify for cleaner assertions
5. Ensure tests are isolated and deterministic

## Coverage Goals

Target coverage levels:
- **Critical paths**: 90%+ coverage
- **Business logic**: 80%+ coverage
- **Overall project**: 70%+ coverage

Check coverage:
```bash
go test ./... -coverprofile=coverage.txt
go tool cover -func=coverage.txt
```
