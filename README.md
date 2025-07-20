# OpenAI Quota Proxy (Go)

OpenAI API proxy server with cost control written in Go.

## Project Structure

```
openai-quota/
├── main.go                    # Main application code
├── main_test.go              # Core functionality tests  
├── additional_test.go        # Advanced tests and edge cases
├── go.mod                    # Go module dependencies
├── go.sum                    # Dependency checksums
├── Makefile                  # Build and test automation
├── README.md                 # Project documentation
├── .gitignore               # Git ignore rules
├── config/                  # Configuration files
│   ├── model_pricing.csv    # OpenAI model pricing data
│   ├── app.env             # Environment configuration template
│   └── README.md           # Configuration documentation
├── scripts/                 # Automation scripts
│   ├── run_tests.sh        # Comprehensive test runner
│   ├── quick_test.sh       # Fast development testing
│   └── README.md           # Script documentation
├── docs/                    # Documentation
│   ├── TESTS.md            # Test suite documentation
│   └── TEST_SCRIPTS.md     # Test script documentation
└── test-reports/            # Generated test reports
```

## Features

- Proxy for OpenAI Chat Completions API
- Control of allowed models 
- Configurable cost quota via command line parameters
- Token counting and cost calculation
- Thread-safe operations
- Real-time cost monitoring
- Dynamic pricing from CSV file
- API key passed via Authorization header
- Support for both `/v1/` and `/api/v1/` endpoints
- Detailed token usage logging

## Usage

### Command line parameters

```bash
# Show available options
./openai-quota -help

# Run with default quota ($2.00)
./openai-quota

# Run with custom quota
./openai-quota -quota 5.0

# Run with custom quota and port
./openai-quota -quota 10.0 -port 8080

# Run with custom pricing file
./openai-quota -quota 5.0 -pricing custom_pricing.csv
```

# Build and run

```bash
# Using Makefile (recommended)
make build
make run

# Custom parameters with Makefile
make run-quota QUOTA=10.0 PORT=8080

# Build manually
go build -o openai-quota

# Run manually
./openai-quota -quota 5.0

# Or directly with go run
go run main.go -quota 5.0 -port 8080
```

### Testing

```bash
# Quick development tests
make test-quick

# Comprehensive test suite
make test-full

# Basic Go tests
make test

# Coverage analysis
make test-coverage
```

## API

### Authorization

All requests must include the OpenAI API key in the Authorization header:

```bash
curl -X POST http://localhost:5000/v1/chat/completions \
     -H "Authorization: Bearer your-openai-api-key" \
     -H "Content-Type: application/json" \
     -d '{"model":"gpt-4o","messages":[{"role":"user","content":"Hello"}]}'
```

### POST /v1/chat/completions or /api/v1/chat/completions

Main proxy endpoint - accepts the same parameters as OpenAI API.

Example response includes additional `proxy_usage` field:

```json
{
  "id": "chatcmpl-...",
  "object": "chat.completion",
  "created": 1234567890,
  "model": "gpt-4o",
  "choices": [...],
  "usage": {
    "prompt_tokens": 10,
    "completion_tokens": 20,
    "total_tokens": 30
  },
  "proxy_usage": {
    "prompt_tokens": 10,
    "completion_tokens": 20,
    "cost_usd": 0.001200
  }
}
```

### GET /v1/chat/completions or /api/v1/chat/completions

Returns server status and current costs:

```json
{
  "info": "Local OpenAI proxy. Available method: POST.",
  "cost_limit": 5.0,
  "current_cost": 1.25,
  "remaining": 3.75,
  "available_models": ["gpt-4o", "gpt-4o-mini", ...],
  "models_count": 23
}
```

### GET /pricing or /api/pricing

Returns detailed pricing information for all models loaded from CSV.

### GET /health

Health check endpoint.

## Application parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `-quota` | Global cost limit in USD | 2.0 |
| `-port` | Server port | 5000 |
| `-pricing` | Path to CSV pricing file | config/model_pricing.csv |
| `-help`, `-h` | Show help | - |

## Configuration

### Model Pricing
Model pricing is loaded from `config/model_pricing.csv`. See `config/README.md` for details on the format and adding new models.

### Environment Variables
Use `config/app.env` as a template for environment configuration.

### Scripts
- `scripts/run_tests.sh` - Comprehensive test suite with detailed reporting
- `scripts/quick_test.sh` - Fast development testing
- See `scripts/README.md` for detailed script documentation

### Documentation
- `docs/TESTS.md` - Test suite documentation  
- `docs/TEST_SCRIPTS.md` - Test script documentation

## Pricing file format

CSV file with model pricing (prices per 1M tokens):

```csv
model,version,input,cached_input,output
gpt-4o,gpt-4o-2024-08-06,2.5,1.25,10.0
gpt-4o-mini,gpt-4o-mini-2024-07-18,0.15,0.075,0.6
```

## Example usage

```bash
# Development server with low quota
./openai-quota -quota 1.0 -port 3000

# Production server with high quota
./openai-quota -quota 50.0 -port 80

# Check current cost status
curl http://localhost:5000/v1/chat/completions

# Make API call
curl -X POST http://localhost:5000/v1/chat/completions \
     -H "Authorization: Bearer sk-your-api-key" \
     -H "Content-Type: application/json" \
     -d '{"model":"gpt-4o","messages":[{"role":"user","content":"Hello world"}]}'
```

## Logging

The server logs detailed information about each request:

```
2025/07/20 10:30:15 Request: model=gpt-4o, prompt_tokens=15, completion_tokens=25, cost=$0.000150, total_cost=$1.250000, remaining=$3.750000
```

## Error responses

- `401 Unauthorized` - Missing or invalid Authorization header
- `400 Bad Request` - Invalid JSON or disallowed model
- `429 Too Many Requests` - Cost limit exceeded
- `500 Internal Server Error` - OpenAI API error
