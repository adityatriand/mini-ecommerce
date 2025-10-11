#!/bin/bash

set -e

echo "Setting up development environment..."

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$PROJECT_ROOT"

echo "Configuring git hooks..."
git config core.hooksPath .githooks
echo "Git hooks configured to use .githooks directory"

echo "Downloading Go dependencies..."
go mod download

echo "Installing development tools..."
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

echo ""
echo "Setup complete!"
echo "Git hooks are now active. Tests will run automatically before commits."
