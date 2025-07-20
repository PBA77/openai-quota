# Test Documentation

## Overview
Comprehensive test suite for the OpenAI Quota Proxy application with **103 individual test cases** covering all major functionality.

## Test Coverage
- **Total Coverage**: 62.4% of code statements
- **Test Files**: 2 files (`main_test.go`, `additional_test.go`)
- **Number of Tests**: 103 individual test cases

## Test Categories

### 1. Unit Tests (Core Functions)

#### CSV Parsing & Model Pricing
- `TestLoadModelPricing` - CSV file loading with various scenarios
- `TestParseFloat` - Float parsing edge cases  
- `TestGetPricingForModel` - Model pricing retrieval and fallbacks
- `TestModelPricing_VersionHandling` - Version-specific model handling

#### Cost Calculation
- `TestCalculateCost` - Token cost calculation accuracy
- `TestCostCalculation_EdgeCases` - Edge cases (zero tokens, large numbers)
- `TestFloatingPointPrecision` - Precision handling

#### Token Management
- `TestCountTokens` - Basic token counting
- `TestCalculateTokensFromMessages` - Message-based token calculation
- `TestTokenCounting_DifferentModels` - Cross-model token counting

#### Model Validation
- `TestIsModelAllowed` - Model whitelist validation
- `TestGetAvailableModels` - Available model listing

### 2. Integration Tests (HTTP Endpoints)

#### Health & Info Endpoints
- `TestHealthEndpoint` - `/health` endpoint functionality
- `TestInfoEndpoint` - `/v1/chat/completions` GET (info)
- `TestPricingEndpoint` - `/pricing` endpoint

#### Authentication & Authorization
- `TestChatCompletionsProxy_MissingAuth` - Missing Authorization header
- `TestChatCompletionsProxy_InvalidAuth` - Invalid Authorization formats
- Multiple auth validation scenarios

#### Request Validation
- `TestChatCompletionsProxy_InvalidJSON` - Malformed JSON handling
- `TestChatCompletionsProxy_DisallowedModel` - Model restriction enforcement
- `TestChatRequest_EmptyMessages` - Empty message handling

#### Quota Management
- `TestChatCompletionsProxy_QuotaExceeded` - Global quota enforcement
- `TestChatCompletionsProxy_PromptCostExceedsQuota` - Preemptive cost blocking

### 3. System Tests

#### Concurrency & Performance
- `TestConcurrentProxyRequests` - Concurrent request handling
- `TestConcurrentAccess` - Thread-safe operations
- `TestPerformanceBaseline` - Performance benchmarks

#### Error Handling
- `TestErrorHandling_InvalidModelInPricing` - Invalid model handling
- `TestErrorResponseFormats` - Error response structure
- `TestFullWorkflow_ValidRequest` - End-to-end request flow

#### Configuration & State
- `TestGlobalStateManagement` - Global variable management
- `TestConfigurationValidation` - Configuration validation
- `TestMemoryUsage` - Memory usage patterns

### 4. Data Format Tests

#### CSV Parsing Edge Cases
- `TestCSVParsing_RealWorldScenarios` - Real-world CSV scenarios
  - Windows line endings
  - Extra whitespace
  - Missing headers
  - Unicode model names

#### JSON Response Validation
- `TestJSONResponseFormats` - Response JSON structure
- `TestValidateStructFields` - Struct field validation

#### HTTP Method Testing
- `TestHTTPMethodsOnEndpoints` - HTTP method enforcement
- GET/POST/PUT method validation per endpoint

### 5. Edge Case Tests

#### Data Edge Cases
- `TestChatRequest_VeryLongContent` - Large request handling
- `TestModelPricing_DeepCopy` - Struct copying behavior

## Test Results Summary

### Passing Tests: ✅ All 103 tests pass
### Coverage Breakdown by Function:
- `loadModelPricing`: 84.4%
- `parseFloat`: 100.0%
- `getPricingForModel`: 100.0%
- `calculateCost`: 100.0%
- `getAvailableModels`: 100.0%
- `isModelAllowed`: 100.0%
- `countTokens`: 100.0%
- `calculateTokensFromMessages`: 100.0%
- `callOpenAI`: 63.6%
- `chatCompletionsProxy`: 72.0%
- `info`: 100.0%
- `pricing`: 100.0%
- `main`: 0.0% (not tested - main function)

### Areas Not Covered by Tests:
1. **Main function** (0.0% coverage) - Entry point, hard to test
2. **Some OpenAI API success paths** - Would require mocking external API
3. **Some error branches in CSV parsing** - Edge cases in file handling

## Running Tests

### All Tests
```bash
go test -v
```

### With Coverage
```bash
go test -cover
```

### Detailed Coverage Report
```bash
go test -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

### Performance Tests Only
```bash
go test -run "Performance|Concurrent" -v
```

## Test Files Structure

### main_test.go
- Core functionality tests
- HTTP endpoint tests  
- Integration tests
- Benchmark tests

### additional_test.go
- Extended edge case tests
- Performance and concurrency tests
- Error handling scenarios
- Configuration validation

### Test Dependencies
- `github.com/gin-gonic/gin` (test mode)
- Standard Go testing package
- `net/http/httptest` for HTTP testing

## Key Testing Strategies Used

1. **Table-Driven Tests** - Multiple test cases per function
2. **HTTP Testing** - Using httptest.ResponseRecorder
3. **Concurrency Testing** - Goroutines and wait groups  
4. **Error Injection** - Invalid data and edge cases
5. **State Management** - Setup/teardown with resetGlobalState()
6. **Performance Testing** - Timing-based benchmarks
7. **Coverage Analysis** - Statement-level coverage tracking

## Quality Metrics

- **Test-to-Code Ratio**: ~2:1 (test files larger than main code)
- **Edge Case Coverage**: Extensive (empty data, invalid formats, unicode, etc.)
- **Error Path Testing**: Comprehensive error scenario coverage
- **Integration Testing**: Full HTTP request/response cycle testing
- **Documentation**: All test functions clearly named and documented
