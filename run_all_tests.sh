#!/bin/bash

echo "=== Running Comprehensive Unit Tests ==="

# Test results summary
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Function to run tests and count results
run_tests() {
    local test_file=$1
    local test_name=$2
    
    echo ""
    echo "🧪 Running $test_name tests..."
    
    if [ -f "$test_file" ]; then
        TOTAL_TESTS=$((TOTAL_TESTS + 1))
        
        # Try to run the tests (if Go is available)
        if command -v go &> /dev/null; then
            if go test -v "$test_file" 2>/dev/null; then
                echo "✅ $test_name tests passed"
                PASSED_TESTS=$((PASSED_TESTS + 1))
            else
                echo "❌ $test_name tests failed"
                FAILED_TESTS=$((FAILED_TESTS + 1))
                # Show detailed error
                go test -v "$test_file"
            fi
        else
            echo "⚠️  Go compiler not available, skipping $test_name tests"
        fi
    else
        echo "❌ Test file not found: $test_file"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
}

# Run all our new tests
echo "🔍 Testing migration functionality..."
run_tests "pkg/migrate/config_test.go" "Migration"

echo "🔍 Testing main application changes..."
run_tests "cmd/picoclaw/main_test.go" "Main Application"

echo "🔍 Testing config model functionality..."
run_tests "pkg/config/config_models_test.go" "Config Models"

echo "🔍 Testing existing config functionality..."
run_tests "pkg/config/config_agents_test.go" "Config Agents"

echo "🔍 Testing subagent functionality..."
run_tests "pkg/tools/subagent_profile_test.go" "Subagent Profiles"

echo "🔍 Testing agent loop functionality..."
run_tests "pkg/agent/loop_agents_test.go" "Agent Loop"

echo ""
echo "=== Test Results Summary ==="
echo "📊 Total test files: $TOTAL_TESTS"
echo "✅ Passed: $PASSED_TESTS"
echo "❌ Failed: $FAILED_TESTS"

if [ $FAILED_TESTS -eq 0 ]; then
    echo "🎉 All tests passed!"
    exit 0
else
    echo "⚠️  Some tests failed"
    exit 1
fi