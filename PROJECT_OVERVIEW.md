# Project Overview

## OpenAI Quota Proxy - Organized Structure

This document provides an overview of the newly organized project structure.

### 📁 Project Structure

```
openai-quota/
├── 🗂️ config/               # Configuration files
│   ├── model_pricing.csv    # OpenAI model pricing data (23+ models)
│   ├── app.env             # Environment template
│   └── README.md           # Configuration documentation
├── 🗂️ scripts/              # Automation scripts  
│   ├── run_tests.sh        # Comprehensive test runner (11 categories)
│   ├── quick_test.sh       # Fast development testing
│   └── README.md           # Script documentation
├── 🗂️ docs/                 # Documentation
│   ├── TESTS.md            # Test suite documentation
│   └── TEST_SCRIPTS.md     # Test script documentation
├── 🗂️ test-reports/         # Generated reports (gitignored)
├── 📄 main.go               # Core application (Go 1.21+)
├── 🧪 main_test.go          # Core functionality tests
├── 🧪 additional_test.go    # Advanced tests (103 total)
├── 🔧 Makefile              # Build automation (15 commands)
├── 📝 README.md             # Project documentation
├── 🚫 .gitignore            # Git ignore rules
├── 📦 go.mod                # Go dependencies
└── 📦 go.sum                # Dependency checksums
```

### ✨ Key Features

#### Core Functionality
- **OpenAI API Proxy** with cost control
- **Thread-safe operations** with mutex protection
- **Preemptive cost blocking** before API calls
- **Token counting** using pkoukk/tiktoken-go
- **Dynamic model pricing** from CSV
- **API key validation** via Authorization header
- **Dual endpoint support** (`/v1/` and `/api/v1/`)

#### Development & Testing
- **103 comprehensive tests** (62.4% coverage)
- **Race condition detection** and safety
- **Performance benchmarking**
- **Automated test reporting** with HTML coverage
- **Quick development testing** (< 30 seconds)
- **CI/CD ready** scripts and automation

#### Configuration Management
- **CLI parameters** for runtime configuration
- **CSV-based pricing** for easy model management
- **Environment templates** for different deployments
- **Comprehensive documentation** for all components

### 🚀 Quick Start

```bash
# Clone and build
git clone <repo>
cd openai-quota
make build

# Run with default settings
make run

# Run with custom parameters
make run-quota QUOTA=10.0 PORT=8080

# Development testing
make test-quick

# Full test suite
make test-full
```

### 🧪 Testing Infrastructure

#### Quick Testing (`make test-quick`)
- Build verification
- Core test execution
- Coverage check
- 103 tests in ~20 seconds

#### Comprehensive Testing (`make test-full`)
- **Basic Tests**: Core functionality
- **Coverage Analysis**: HTML reports
- **Race Detection**: Concurrency safety
- **Performance Tests**: Benchmark validation
- **Security Tests**: Authentication & validation
- **Error Handling**: Edge cases
- **Integration Tests**: End-to-end API
- **Static Analysis**: Code quality
- **Build Verification**: Multi-platform
- **Documentation**: Completeness checks
- **Configuration**: Parameter validation

### 📊 Test Coverage

```
Total Tests: 103
Coverage: 62.4%
Race Detection: ✅ PASS
Performance: ✅ PASS
Security: ✅ PASS
```

### 🛠️ Available Make Commands

```bash
make build              # Build application
make run                # Run with defaults
make run-quota          # Run with custom params
make test               # Basic Go tests
make test-quick         # Fast development testing
make test-full          # Comprehensive test suite
make test-coverage      # Coverage analysis
make clean              # Clean build artifacts
make help               # Show all commands
```

### 📝 Configuration

#### Model Pricing (`config/model_pricing.csv`)
- 23+ OpenAI models supported
- Easy to update with new pricing
- Automatic loading at startup

#### Environment (`config/app.env`)
- Production/development templates
- Security and performance settings
- Logging configuration

### 🔒 Security Features

- API key validation from Authorization header
- Request size limits
- Rate limiting support
- Input sanitization
- Error message sanitization

### 📈 Performance Features

- Thread-safe global state management
- Efficient token counting with caching
- Preemptive cost calculation
- Background process support
- Configurable timeouts

### 🚦 CI/CD Ready

- Comprehensive test automation
- HTML coverage reports
- Race condition detection
- Multi-platform build support
- Detailed logging and monitoring

---

**Status**: ✅ Production Ready  
**Tests**: 103 passing (62.4% coverage)  
**Go Version**: 1.21+  
**Dependencies**: Minimal (Gin, tiktoken-go)  
**Maintenance**: Automated testing & reporting
