# Scripts Directory

This directory contains automation scripts for the OpenAI Quota Proxy project.

## Available Scripts

### `run_tests.sh`
Comprehensive test runner that executes all test categories:

- **Basic Tests**: Core functionality testing
- **Coverage Analysis**: Code coverage reporting with HTML output
- **Race Detection**: Concurrency safety validation
- **Performance Tests**: Benchmark testing
- **Security Tests**: Authentication and validation checks
- **Error Handling**: Error scenarios and edge cases
- **Integration Tests**: End-to-end API testing
- **Static Analysis**: Code quality checks (go vet, go fmt)
- **Build Verification**: Compilation and build testing
- **Documentation**: README and docs validation
- **Configuration**: Config file and CLI parameter testing

**Usage:**
```bash
./scripts/run_tests.sh
```

**Features:**
- Detailed HTML coverage report
- Test categorization and timing
- Colored output for easy reading
- Comprehensive test summary
- Pass/fail statistics

### `quick_test.sh`
Fast development testing script for rapid feedback:

- Build verification
- Core tests execution
- Coverage check
- Quick summary

**Usage:**
```bash
./scripts/quick_test.sh
```

**Features:**
- Fast execution (under 30 seconds)
- Essential test coverage
- Development-focused output
- Build verification

## Running Scripts

From project root directory:

```bash
# Full comprehensive testing
make test-full

# Quick development testing
make test-quick

# Or run directly
./scripts/run_tests.sh
./scripts/quick_test.sh
```

## Script Output

Both scripts provide:
- Build status verification
- Test execution results
- Coverage statistics
- Pass/fail summary
- Timing information

The comprehensive test runner additionally provides:
- Detailed HTML coverage report in `test-reports/`
- Test categorization
- Static analysis results
- Race condition detection
- Performance benchmarks
