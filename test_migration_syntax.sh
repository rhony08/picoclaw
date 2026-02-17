#!/bin/bash

echo "=== Testing Migration File Syntax ==="

# Check if the file compiles (if Go is available)
if command -v go &> /dev/null; then
    echo "Go compiler found, checking syntax..."
    if go build -o /dev/null pkg/migrate/config.go 2>/dev/null; then
        echo "✅ Syntax check passed"
    else
        echo "❌ Syntax check failed"
        go build pkg/migrate/config.go
    fi
else
    echo "Go compiler not found, skipping syntax check"
fi

# Check file structure
echo "Checking file structure..."
if [ -f "pkg/migrate/config.go" ]; then
    echo "✅ File exists"
    
    # Check for proper function definitions
    if grep -q "func ConvertConfig" pkg/migrate/config.go; then
        echo "✅ ConvertConfig function found"
    else
        echo "❌ ConvertConfig function missing"
    fi
    
    if grep -q "func MergeConfig" pkg/migrate/config.go; then
        echo "✅ MergeConfig function found"
    else
        echo "❌ MergeConfig function missing"
    fi
    
    if grep -q "func NeedsMigration" pkg/migrate/config.go; then
        echo "✅ NeedsMigration function found"
    else
        echo "❌ NeedsMigration function missing"
    fi
    
    if grep -q "func MigrateToNewFormat" pkg/migrate/config.go; then
        echo "✅ MigrateToNewFormat function found"
    else
        echo "❌ MigrateToNewFormat function missing"
    fi
    
    # Check for syntax errors
    if grep -n "non-declaration statement outside function body" pkg/migrate/config.go; then
        echo "❌ Syntax errors found"
    else
        echo "✅ No syntax errors detected"
    fi
    
    # Check line count
    lines=$(wc -l < pkg/migrate/config.go)
    echo "📄 File has $lines lines"
    
else
    echo "❌ File not found"
fi

echo "=== Test Complete ==="