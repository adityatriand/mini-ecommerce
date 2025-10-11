#!/bin/bash

set -e

echo "Running pre-commit tests..."

STAGED_GO_FILES=$(git diff --cached --name-only --diff-filter=ACM | grep '\.go$' || true)

if [ -z "$STAGED_GO_FILES" ]; then
    echo "No Go files staged, skipping tests"
    exit 0
fi

PACKAGES_TO_TEST=""

for file in $STAGED_GO_FILES; do
    if [[ $file == internal/auth/* ]]; then
        if [[ ! $PACKAGES_TO_TEST =~ "internal/auth" ]]; then
            PACKAGES_TO_TEST="$PACKAGES_TO_TEST ./internal/auth"
        fi
    elif [[ $file == internal/product/* ]]; then
        if [[ ! $PACKAGES_TO_TEST =~ "internal/product" ]]; then
            PACKAGES_TO_TEST="$PACKAGES_TO_TEST ./internal/product"
        fi
    elif [[ $file == internal/order/* ]]; then
        if [[ ! $PACKAGES_TO_TEST =~ "internal/order" ]]; then
            PACKAGES_TO_TEST="$PACKAGES_TO_TEST ./internal/order"
        fi
    fi
done

if [ -z "$PACKAGES_TO_TEST" ]; then
    echo "No testable packages affected by changes"
    exit 0
fi

echo "Running tests for packages:$PACKAGES_TO_TEST"

for package in $PACKAGES_TO_TEST; do
    echo "Testing $package..."
    if ! go test $package -v; then
        echo "Tests failed for $package"
        echo "Commit aborted. Please fix the tests before committing."
        exit 1
    fi
done

echo "All tests passed!"
exit 0
