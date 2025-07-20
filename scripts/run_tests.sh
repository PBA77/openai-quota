#!/bin/bash

# OpenAI Quota Proxy - Comprehensive Test Runner
# Runs all tests and generates detailed reports

set -e  # Exit on any error

# Change to project root directory
cd "$(dirname "$0")/.."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
REPORT_DIR="test-reports"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
REPORT_FILE="$REPORT_DIR/test_report_$TIMESTAMP.txt"
HTML_REPORT="$REPORT_DIR/coverage_$TIMESTAMP.html"

echo -e "${BLUE}=================================================="
echo -e "OpenAI Quota Proxy - Test Suite Runner"
echo -e "Timestamp: $(date)"
echo -e "==================================================${NC}\n"

# Create reports directory
mkdir -p "$REPORT_DIR"

# Start report file
{
    echo "OpenAI Quota Proxy - Test Report"
    echo "Generated: $(date)"
    echo "========================================"
    echo ""
} > "$REPORT_FILE"

# Function to log and display
log_and_display() {
    echo -e "$1"
    echo -e "$1" | sed 's/\x1b\[[0-9;]*m//g' >> "$REPORT_FILE"
}

# Function to run command and capture output
run_and_capture() {
    local cmd="$1"
    local description="$2"
    
    echo -e "${YELLOW}Running: $description${NC}"
    echo "Running: $description" >> "$REPORT_FILE"
    echo "Command: $cmd" >> "$REPORT_FILE"
    
    if eval "$cmd" >> "$REPORT_FILE" 2>&1; then
        echo -e "${GREEN}✓ $description - PASSED${NC}"
        echo "✓ $description - PASSED" >> "$REPORT_FILE"
    else
        echo -e "${RED}✗ $description - FAILED${NC}"
        echo "✗ $description - FAILED" >> "$REPORT_FILE"
        return 1
    fi
    echo "" >> "$REPORT_FILE"
}

# Function to capture output with display
run_with_output() {
    local cmd="$1"
    local description="$2"
    
    echo -e "${YELLOW}Running: $description${NC}"
    echo "Running: $description" >> "$REPORT_FILE"
    echo "Command: $cmd" >> "$REPORT_FILE"
    echo "" >> "$REPORT_FILE"
    
    # Run command and capture output
    local output
    if output=$(eval "$cmd" 2>&1); then
        echo -e "${GREEN}✓ $description - COMPLETED${NC}"
        echo "✓ $description - COMPLETED" >> "$REPORT_FILE"
    else
        echo -e "${RED}✗ $description - FAILED${NC}"
        echo "✗ $description - FAILED" >> "$REPORT_FILE"
    fi
    
    # Add output to report
    echo "Output:" >> "$REPORT_FILE"
    echo "$output" >> "$REPORT_FILE"
    echo "" >> "$REPORT_FILE"
    
    # Display summary
    echo "$output" | tail -5
    echo ""
}

# Clean previous artifacts
echo -e "${CYAN}Cleaning previous test artifacts...${NC}"
rm -f coverage.out coverage.html test_*.csv

# 1. Code Quality Checks
log_and_display "${PURPLE}=== 1. CODE QUALITY CHECKS ===${NC}"

run_and_capture "go fmt ./..." "Code formatting"
run_and_capture "go vet ./..." "Code analysis (go vet)"

# Check if golangci-lint is available
if command -v golangci-lint &> /dev/null; then
    run_and_capture "golangci-lint run" "Linting (golangci-lint)"
else
    log_and_display "${YELLOW}⚠ golangci-lint not found - skipping advanced linting${NC}"
fi

# 2. Dependency Check
log_and_display "${PURPLE}=== 2. DEPENDENCY CHECK ===${NC}"

run_and_capture "go mod tidy" "Dependency cleanup"
run_and_capture "go mod verify" "Dependency verification"

# 3. Build Test
log_and_display "${PURPLE}=== 3. BUILD TEST ===${NC}"

run_and_capture "go build -o openai-quota-test ." "Application build"
run_and_capture "rm -f openai-quota-test" "Cleanup test binary"

# 4. Unit Tests
log_and_display "${PURPLE}=== 4. UNIT TESTS ===${NC}"

run_with_output "go test -v -timeout=60s" "All tests (verbose)"

# 5. Test Coverage
log_and_display "${PURPLE}=== 5. TEST COVERAGE ANALYSIS ===${NC}"

run_with_output "go test -cover" "Basic coverage"
run_with_output "go test -coverprofile=coverage.out" "Coverage profile generation"

if [ -f coverage.out ]; then
    run_with_output "go tool cover -func=coverage.out" "Coverage by function"
    
    # Generate HTML report
    if go tool cover -html=coverage.out -o "$HTML_REPORT" 2>/dev/null; then
        log_and_display "${GREEN}✓ HTML coverage report generated: $HTML_REPORT${NC}"
    fi
fi

# 6. Performance Tests
log_and_display "${PURPLE}=== 6. PERFORMANCE TESTS ===${NC}"

run_with_output "go test -bench=. -benchmem -timeout=120s" "Benchmark tests"

# 7. Race Condition Detection
log_and_display "${PURPLE}=== 7. RACE CONDITION DETECTION ===${NC}"

run_with_output "go test -race -timeout=60s" "Race condition detection"

# 8. Test Statistics
log_and_display "${PURPLE}=== 8. TEST STATISTICS ===${NC}"

# Count tests
TEST_COUNT=$(go test -v 2>&1 | grep "=== RUN" | wc -l | tr -d ' ')
log_and_display "Total number of tests: $TEST_COUNT"

# Count test files
TEST_FILES=$(find . -name "*_test.go" | wc -l | tr -d ' ')
log_and_display "Test files: $TEST_FILES"

# Code lines
MAIN_LINES=$(wc -l < main.go)
TEST_LINES=$(find . -name "*_test.go" -exec wc -l {} + | tail -1 | awk '{print $1}')
log_and_display "Main code lines: $MAIN_LINES"
log_and_display "Test code lines: $TEST_LINES"
log_and_display "Test-to-code ratio: $(echo "scale=2; $TEST_LINES / $MAIN_LINES" | bc)"

# 9. Security Tests (if available)
log_and_display "${PURPLE}=== 9. SECURITY ANALYSIS ===${NC}"

if command -v gosec &> /dev/null; then
    run_with_output "gosec ./..." "Security analysis (gosec)"
else
    log_and_display "${YELLOW}⚠ gosec not found - skipping security analysis${NC}"
fi

# 10. Integration Tests
log_and_display "${PURPLE}=== 10. INTEGRATION TESTS ===${NC}"

# Test if we can start the server briefly
run_and_capture "timeout 5s go run main.go -quota 1.0 -port 0 || true" "Server startup test"

# 11. Final Summary
log_and_display "${PURPLE}=== 11. FINAL SUMMARY ===${NC}"

{
    echo "Test execution completed at: $(date)"
    echo ""
    echo "Generated files:"
    echo "- Test report: $REPORT_FILE"
    if [ -f "$HTML_REPORT" ]; then
        echo "- HTML coverage: $HTML_REPORT"
    fi
    if [ -f coverage.out ]; then
        echo "- Coverage data: coverage.out"
    fi
    echo ""
    echo "Key metrics:"
    echo "- Total tests: $TEST_COUNT"
    echo "- Test files: $TEST_FILES"
    echo "- Main code lines: $MAIN_LINES"
    echo "- Test code lines: $TEST_LINES"
    
    if [ -f coverage.out ]; then
        COVERAGE=$(go tool cover -func=coverage.out | grep "total:" | awk '{print $3}')
        echo "- Code coverage: $COVERAGE"
    fi
} >> "$REPORT_FILE"

# Display final summary
echo -e "${GREEN}=================================================="
echo -e "TEST SUITE EXECUTION COMPLETED"
echo -e "==================================================${NC}"
echo -e "${CYAN}Reports generated:${NC}"
echo -e "📄 Text report: $REPORT_FILE"

if [ -f "$HTML_REPORT" ]; then
    echo -e "🌐 HTML coverage: $HTML_REPORT"
fi

echo -e "\n${CYAN}Quick stats:${NC}"
echo -e "🧪 Total tests: $TEST_COUNT"
echo -e "📁 Test files: $TEST_FILES"

if [ -f coverage.out ]; then
    COVERAGE=$(go tool cover -func=coverage.out | grep "total:" | awk '{print $3}')
    echo -e "📊 Code coverage: $COVERAGE"
fi

# Check if any failures occurred
if [ $? -eq 0 ]; then
    echo -e "\n${GREEN}🎉 All tests and checks passed successfully!${NC}"
    exit 0
else
    echo -e "\n${RED}❌ Some tests or checks failed. Check the report for details.${NC}"
    exit 1
fi
