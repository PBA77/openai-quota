#!/bin/bash

# Quick Test Runner - OpenAI Quota Proxy
# Simplified version for daily development

set -e

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

#!/bin/bash

# Quick Test Script for OpenAI Quota Proxy
# Fast development testing without full reports

set -e

# Change to project root directory
cd "$(dirname "$0")/.."

echo "🚀 Quick Test Suite"
echo -e "===================="

# 1. Quick build test
echo -e "${YELLOW}🔨 Building...${NC}"
if go build -o /tmp/test-build . && rm -f /tmp/test-build; then
    echo -e "${GREEN}✓ Build successful${NC}"
else
    echo -e "${RED}✗ Build failed${NC}"
    exit 1
fi

# 2. Run tests
echo -e "\n${YELLOW}🧪 Running tests...${NC}"
if go test -timeout=30s; then
    echo -e "${GREEN}✓ All tests passed${NC}"
else
    echo -e "${RED}✗ Tests failed${NC}"
    exit 1
fi

# 3. Quick coverage
echo -e "\n${YELLOW}📊 Coverage check...${NC}"
COVERAGE=$(go test -cover 2>/dev/null | grep "coverage:" | awk '{print $5}' | head -1)
if [ -z "$COVERAGE" ]; then
    COVERAGE=$(go test -cover 2>/dev/null | grep "coverage:" | awk '{print $4}' | head -1)
fi
echo -e "${GREEN}✓ Coverage: ${COVERAGE:-"N/A"}${NC}"

# 4. Test count
TEST_COUNT=$(go test -v 2>&1 | grep "=== RUN" | wc -l | tr -d ' ')
echo -e "${GREEN}✓ Total tests: $TEST_COUNT${NC}"

echo -e "\n${GREEN}🎉 All checks passed!${NC}"
echo -e "For detailed report run: ${YELLOW}./scripts/run_tests.sh${NC}"
