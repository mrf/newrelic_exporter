# Testing Guide

This document describes the testing approach and how to run tests for the New Relic Exporter.

## Test Coverage

The project includes comprehensive unit tests for all major components:

### Config Package (`config/`)
- **config_test.go**: Configuration validation and parsing tests
  - Tests for missing required fields
  - Tests for duration format validation
  - Tests for default value assignment
  - Tests for invalid YAML handling

### Exporter Package (`exporter/`)
- **exporter_test.go**: Prometheus exporter functionality tests
  - Exporter initialization
  - Metric channel communication
  - Prometheus descriptor generation
  - Metric namespacing
  - Cache timing behavior
  - Metric value conversion

### NewRelic Package (`newrelic/`)
- **newrelic_test.go**: New Relic API client tests
  - API initialization
  - Application list retrieval
  - Metric name retrieval
  - Metric data retrieval
  - Authentication/authorization handling
  - Rate limiting (429 responses)
  - Pagination handling
  - Request chunking for large metric sets

### Main Package
- **newrelic_exporter_test.go**: Integration tests
  - End-to-end API interaction
  - Full scrape workflow

## Running Tests

### Run All Tests

```bash
go test ./...
```

### Run Tests with Verbose Output

```bash
go test -v ./...
```

### Run Tests with Coverage

```bash
go test -v -cover ./...
```

### Generate Coverage Report

```bash
# Generate coverage profile
go test -coverprofile=coverage.out ./...

# View coverage in terminal
go tool cover -func=coverage.out

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html
```

### Run Tests for Specific Package

```bash
# Config tests only
go test -v ./config/

# Exporter tests only
go test -v ./exporter/

# NewRelic API tests only
go test -v ./newrelic/
```

### Run Specific Test

```bash
go test -v -run TestConfigValidation ./config/
go test -v -run TestNewExporter ./exporter/
go test -v -run TestGetApplications ./newrelic/
```

### Run Tests with Race Detection

```bash
go test -race ./...
```

This is particularly important for this project since it uses goroutines and channels extensively.

## Test Structure

### Unit Tests

Unit tests focus on individual functions and components in isolation:

```go
func TestFunctionName(t *testing.T) {
    // Setup
    input := "test input"

    // Execute
    result := FunctionToTest(input)

    // Assert
    if result != expected {
        t.Errorf("Expected %v, got %v", expected, result)
    }
}
```

### Integration Tests

Integration tests use mock HTTP servers to test API interactions:

```go
func TestAPIIntegration(t *testing.T) {
    // Create test server
    ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Mock API response
        json.NewEncoder(w).Encode(mockData)
    }))
    defer ts.Close()

    // Test API client against mock server
    api := NewAPI(config)
    result, err := api.GetData()

    // Assertions
    if err != nil {
        t.Fatal(err)
    }
}
```

## Test Data

Test data files are located in the `_testing/` directory:

- `application_list.json` - Mock application list response
- `metric_names.json` - Mock metric names response
- `metric_data.json` - Mock metric data response

These files are used by integration tests to simulate New Relic API responses.

## Continuous Integration

Tests are automatically run on:

- **Every push** to any branch
- **Every pull request**
- **Before releases**

CI systems used:
- GitHub Actions (see `.github/workflows/`)
- CircleCI (see `.circleci/config.yml`)

Both CI systems run:
1. `go test` with race detection
2. `go vet` for static analysis
3. `gofmt` for code formatting checks
4. Coverage report generation

## Writing New Tests

When adding new functionality, include tests that cover:

1. **Happy path** - Normal, expected usage
2. **Error cases** - Invalid inputs, API errors
3. **Edge cases** - Empty data, large datasets, boundary conditions
4. **Concurrency** - If using goroutines, test with `-race` flag

### Test Naming Conventions

- Test files: `*_test.go`
- Test functions: `TestFunctionName`
- Benchmark functions: `BenchmarkFunctionName`
- Example functions: `ExampleFunctionName`

### Best Practices

1. **Isolation**: Tests should not depend on external services
2. **Independence**: Tests should not depend on each other
3. **Clarity**: Test names should clearly describe what is being tested
4. **Completeness**: Test both success and failure scenarios
5. **Speed**: Keep tests fast by using mocks instead of real API calls

## Code Coverage Goals

Target code coverage metrics:

- **Overall**: > 70%
- **Critical paths**: > 90% (config validation, API requests)
- **New code**: 100% of new functionality should have tests

## Benchmarking

Run benchmarks to test performance:

```bash
# Run all benchmarks
go test -bench=. ./...

# Run specific benchmark
go test -bench=BenchmarkName ./package/

# With memory profiling
go test -bench=. -benchmem ./...
```

## Debugging Tests

### Verbose Output

```bash
go test -v ./...
```

### Run Single Test with Logging

```bash
go test -v -run TestName ./package/ 2>&1 | less
```

### Debug with Delve

```bash
dlv test ./package/
```

## Common Issues

### Test Timeouts

If tests are timing out:

```bash
go test -timeout 30s ./...
```

### Import Cycles

If you encounter import cycle errors, restructure packages to remove circular dependencies.

### Race Conditions

Always run tests with race detection before committing:

```bash
go test -race ./...
```

## Additional Resources

- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [Go Test Command](https://golang.org/cmd/go/#hdr-Test_packages)
- [Table-Driven Tests in Go](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests)
- [Testing with Mock HTTP Servers](https://golang.org/pkg/net/http/httptest/)

## Questions?

If you have questions about testing:
- Check existing tests for examples
- Review this guide
- Ask in GitHub Issues or Discussions
