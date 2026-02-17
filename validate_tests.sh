#!/bin/bash

echo "=== Comprehensive Test Validation ==="

# Test validation results
TOTAL_CHECKS=0
PASSED_CHECKS=0
FAILED_CHECKS=0

# Function to validate test files
validate_test_file() {
    local test_file=$1
    local test_name=$2
    
    echo ""
    echo "🔍 Validating $test_name..."
    
    if [ ! -f "$test_file" ]; then
        echo "❌ Test file not found: $test_file"
        FAILED_CHECKS=$((FAILED_CHECKS + 1))
        return
    fi
    
    TOTAL_CHECKS=$((TOTAL_CHECKS + 1))
    
    # Check if file has proper test structure
    if grep -q "func Test" "$test_file"; then
        echo "✅ Has test functions"
        PASSED_CHECKS=$((PASSED_CHECKS + 1))
    else
        echo "❌ Missing test functions"
        FAILED_CHECKS=$((FAILED_CHECKS + 1))
    fi
    
    # Check for imports
    if grep -q "import" "$test_file"; then
        echo "✅ Has imports"
        PASSED_CHECKS=$((PASSED_CHECKS + 1))
    else
        echo "❌ Missing imports"
        FAILED_CHECKS=$((FAILED_CHECKS + 1))
    fi
    
    # Check for t.Run usage
    if grep -q "t\.Run" "$test_file"; then
        echo "✅ Uses subtests"
        PASSED_CHECKS=$((PASSED_CHECKS + 1))
    else
        echo "⚠️  Doesn't use subtests (optional)"
    fi
    
    # Check for t.Error/t.Errorf usage
    if grep -q "t\.Error" "$test_file"; then
        echo "✅ Has error assertions"
        PASSED_CHECKS=$((PASSED_CHECKS + 1))
    else
        echo "❌ Missing error assertions"
        FAILED_CHECKS=$((FAILED_CHECKS + 1))
    fi
    
    # Count test functions
    test_count=$(grep -c "func Test" "$test_file")
    echo "📊 Test functions: $test_count"
    
    # Show test function names
    echo "📋 Test functions:"
    grep -o "func Test[^(]*" "$test_file" | sed 's/func Test/  - /'
}

# Validate all our test files
echo "🧪 Validating migration tests..."
validate_test_file "pkg/migrate/config_test.go" "Migration"

echo "🧪 Validating main application tests..."
validate_test_file "cmd/picoclaw/main_test.go" "Main Application"

echo "🧪 Validating config model tests..."
validate_test_file "pkg/config/config_models_test.go" "Config Models"

echo "🧪 Validating existing config tests..."
validate_test_file "pkg/config/config_agents_test.go" "Config Agents"

echo "🧪 Validating subagent tests..."
validate_test_file "pkg/tools/subagent_profile_test.go" "Subagent Profiles"

echo "🧪 Validating agent loop tests..."
validate_test_file "pkg/agent/loop_agents_test.go" "Agent Loop"

echo ""
echo "=== Validation Results ==="
echo "📊 Total checks: $TOTAL_CHECKS"
echo "✅ Passed: $PASSED_CHECKS"
echo "❌ Failed: $FAILED_CHECKS"

if [ $FAILED_CHECKS -eq 0 ]; then
    echo "🎉 All validations passed!"
    exit 0
else
    echo "⚠️  Some validations failed"
    exit 1
fi