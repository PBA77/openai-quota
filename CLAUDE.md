# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Build and Run
```bash
# Build the application
make build

# Run with default settings (quota=$2.00, port=5000)
make run

# Run with custom parameters
make run-quota QUOTA=5.0 PORT=8080

# Run compiled binary
make run-binary
```

### Testing
```bash
# Quick development tests (~20 seconds, 103 tests)
make test-quick

# Comprehensive test suite with reports
make test-full

# Basic Go tests
make test

# Code coverage analysis
make test-coverage

# HTML coverage report
make test-html

# Race condition detection
go test -race ./...
```

### Code Quality
```bash
# Format code
make fmt

# Static analysis
make vet

# Clean build artifacts
make clean
```

## Architecture Overview

This is a Go-based OpenAI API proxy server with cost control features. The application is organized as a single-file main.go with comprehensive testing.

### Core Components

**Main Application (main.go)**
- Single-file architecture using Gin web framework
- Thread-safe cost tracking with mutex protection
- Token counting using tiktoken-go library
- Dynamic model pricing loaded from CSV
- Support for both `/v1/` and `/api/v1/` endpoints

**Key Data Structures:**
- `ModelPricing`: Defines pricing for OpenAI models (input, cached_input, output per 1M tokens)
- `ChatRequest`/`ChatResponse`: OpenAI API request/response structures with proxy extensions
- Global variables: `costLimitUSD`, `totalCost`, `modelPricing` map, `mu` mutex

### Configuration Management

**Model Pricing (`config/model_pricing.csv`)**
- CSV format: `model,version,input,cached_input,output`
- Prices per 1M tokens for 23+ OpenAI models
- Loaded at startup into `modelPricing` map

**CLI Parameters:**
- `-quota`: Cost limit in USD (default: 2.0)
- `-port`: Server port (default: 5000)  
- `-pricing`: Path to CSV pricing file

### API Endpoints

- `POST /v1/chat/completions` or `/api/v1/chat/completions`: Main proxy endpoint
- `GET /v1/chat/completions` or `/api/v1/chat/completions`: Server status and costs
- `GET /pricing` or `/api/pricing`: Model pricing information
- `GET /health`: Health check

### Security Features

- API key validation via Authorization header
- Model allowlist with predefined prefixes
- Preemptive cost checking before OpenAI API calls
- Request size limits and input sanitization

### Testing Infrastructure

The project has 103 comprehensive tests across:
- Core functionality (main_test.go)
- Advanced scenarios (additional_test.go)
- Race condition safety
- Performance benchmarks
- Security validation
- Error handling edge cases

**Test Scripts:**
- `scripts/quick_test.sh`: Fast development testing
- `scripts/run_tests.sh`: Full test suite with 11 categories of tests

## Dependencies

**Required:**
- Go 1.21+
- github.com/gin-gonic/gin v1.9.1 (web framework)
- github.com/pkoukk/tiktoken-go v0.1.6 (token counting)

**Development Tools:**
- Comprehensive Makefile with 15+ commands
- Test automation scripts
- HTML coverage reporting
- Multi-platform build support

## Important Notes

- Single-file architecture keeps all logic in main.go
- Thread-safety is critical - all cost updates use mutex protection
- Cost checking happens BEFORE making OpenAI API calls to prevent overruns
- Model pricing is loaded once at startup from CSV file
- The application logs detailed request information including costs
- Test coverage is 62.4% with focus on critical paths and edge cases