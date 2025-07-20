# Test Scripts Documentation

This directory contains comprehensive testing scripts for the OpenAI Quota Proxy project.

## Available Scripts

### 1. 🚀 `./run_tests.sh` - Comprehensive Test Suite

**Complete test runner with detailed reporting**

Features:
- ✅ Code quality checks (formatting, vetting, linting)
- ✅ Dependency verification
- ✅ Build testing  
- ✅ All unit tests with verbose output
- ✅ Test coverage analysis with HTML report
- ✅ Performance benchmarks
- ✅ Race condition detection
- ✅ Security analysis (if tools available)
- ✅ Integration tests
- ✅ Detailed statistics and reporting

**Usage:**
```bash
./run_tests.sh
```

**Generates:**
- `test-reports/test_report_YYYYMMDD_HHMMSS.txt` - Detailed text report
- `test-reports/coverage_YYYYMMDD_HHMMSS.html` - HTML coverage report

**Time:** ~1-2 minutes

---

### 2. ⚡ `./quick_test.sh` - Fast Development Testing

**Quick test runner for daily development**

Features:
- ✅ Fast build check
- ✅ All tests execution
- ✅ Basic coverage info
- ✅ Test count summary

**Usage:**
```bash
./quick_test.sh
```

**Time:** ~5-15 seconds

---

## Makefile Integration

You can also run tests through make commands:

### Basic Testing
```bash
make test                    # Run all tests
make test-verbose           # Run tests with detailed output
make test-coverage          # Run tests with coverage
make test-html              # Generate HTML coverage report
```

### Advanced Testing
```bash
make test-bench             # Run benchmark tests
make test-count             # Count total tests
make test-coverage-func     # Show coverage per function
make test-quick             # Run quick test script
make test-full-report       # Run comprehensive test script
```

### Development Workflow
```bash
make test-run               # Run specific test pattern
make clean                  # Clean all test artifacts
```

---

## Test Statistics

- **Total Tests:** 103 individual test cases
- **Test Files:** 2 (`main_test.go`, `additional_test.go`)
- **Code Coverage:** 62.4%
- **Test Categories:** Unit, Integration, System, Performance, Security

### Coverage Breakdown
- `parseFloat`: 100.0%
- `getPricingForModel`: 100.0%
- `calculateCost`: 100.0%
- `getAvailableModels`: 100.0%
- `isModelAllowed`: 100.0%
- `countTokens`: 100.0%
- `calculateTokensFromMessages`: 100.0%
- `info`: 100.0%
- `pricing`: 100.0%
- `loadModelPricing`: 84.4%
- `chatCompletionsProxy`: 72.0%
- `callOpenAI`: 63.6%
- `main`: 0.0% (entry point)

---

## Test Categories

### 1. Unit Tests
- CSV parsing and model pricing
- Cost calculation and token management
- Model validation and whitelisting
- Error handling and edge cases

### 2. Integration Tests  
- HTTP endpoint testing (`/health`, `/pricing`, `/v1/chat/completions`)
- Authentication and authorization
- Request validation and error responses
- Quota management and blocking

### 3. System Tests
- Concurrency and thread safety
- Performance benchmarks
- Memory usage and state management
- Configuration validation

### 4. Edge Case Tests
- Unicode and special characters
- Large requests and malformed data
- Network error simulation
- Resource exhaustion scenarios

---

## Continuous Integration

For CI/CD pipelines, use the comprehensive test script:

```yaml
# Example GitHub Actions
- name: Run comprehensive tests
  run: ./run_tests.sh
```

```dockerfile
# Example Dockerfile testing
RUN chmod +x run_tests.sh && ./run_tests.sh
```

---

## Prerequisites

### Required Tools
- Go 1.21+ (for running tests)
- `bc` (for calculations in scripts)

### Optional Tools (Enhanced Features)
- `golangci-lint` (advanced linting)
- `gosec` (security analysis)
- `entr` (watch mode for development)

### Installation
```bash
# Install optional tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
```

---

## Troubleshooting

### Common Issues

**Permission denied:**
```bash
chmod +x run_tests.sh quick_test.sh
```

**Missing bc calculator:**
```bash
# macOS
brew install bc

# Ubuntu/Debian  
sudo apt-get install bc
```

**Coverage report not generated:**
```bash
# Ensure go tool cover is available
go tool cover -h
```

### Performance Notes

- **Comprehensive tests** may be slow due to tiktoken library initialization
- **Race detection** adds significant overhead
- **Benchmark tests** require multiple iterations for accuracy

---

## Contributing

When adding new tests:

1. Add unit tests to `main_test.go` 
2. Add advanced tests to `additional_test.go`
3. Update test count in documentation
4. Run `./run_tests.sh` to verify everything works

### Test Naming Convention
- `Test[Function]` - Basic functionality
- `Test[Function]_EdgeCases` - Edge cases and error conditions  
- `Test[Area]_[Scenario]` - Integration and system tests

---

## Reports Archive

Test reports are automatically saved in `test-reports/` directory:
- Text reports: `test_report_YYYYMMDD_HHMMSS.txt`
- HTML coverage: `coverage_YYYYMMDD_HHMMSS.html`

Clean old reports with:
```bash
make clean
# or
rm -rf test-reports/
```
