#!/bin/sh
# ClaraVerse Backend Test Runner
# This script runs all unit tests and integration tests
# Exit code 0 = all tests passed, non-zero = failure

set -e

echo "=========================================="
echo "  ClaraVerse Backend Test Suite"
echo "=========================================="
echo ""

# Track test failures
FAILURES=0

# Function to run tests for a package
run_package_tests() {
    local package=$1
    local description=$2

    echo "----------------------------------------"
    echo "Testing: $description"
    echo "Package: $package"
    echo "----------------------------------------"

    if go test -v -race -timeout 60s "$package"; then
        echo "✅ PASSED: $description"
    else
        echo "❌ FAILED: $description"
        FAILURES=$((FAILURES + 1))
    fi
    echo ""
}

# Core packages
echo "=== Core Package Tests ==="
run_package_tests "./internal/database" "Database"
run_package_tests "./internal/models/..." "Models"

# Service tests
echo "=== Service Tests ==="
run_package_tests "./internal/services/..." "Services"

# Handler tests
echo "=== Handler Tests ==="
run_package_tests "./internal/handlers/..." "Handlers"

# Tool tests (includes all file tools)
echo "=== Tool Tests ==="
run_package_tests "./internal/tools/..." "Tools"

# Execution tests (workflow, variable executor)
echo "=== Execution Tests ==="
run_package_tests "./internal/execution/..." "Execution"

# File cache tests
echo "=== File Cache Tests ==="
run_package_tests "./internal/filecache/..." "File Cache"

# Audio service tests
echo "=== Audio Service Tests ==="
run_package_tests "./internal/audio/..." "Audio Service"

# Vision service tests
echo "=== Vision Service Tests ==="
run_package_tests "./internal/vision/..." "Vision Service"

# Preflight tests
echo "=== Preflight Tests ==="
run_package_tests "./internal/preflight/..." "Preflight Checks"

# Integration tests
echo "=== Integration Tests ==="
run_package_tests "./tests/..." "Integration Tests"

# Summary
echo "=========================================="
echo "  Test Summary"
echo "=========================================="

if [ $FAILURES -eq 0 ]; then
    echo "✅ ALL TESTS PASSED"
    echo ""
    exit 0
else
    echo "❌ $FAILURES TEST SUITE(S) FAILED"
    echo ""
    exit 1
fi
