# Git Hooks

This directory contains git hooks to ensure code quality before commits.

## Pre-commit Hook

The pre-commit hook automatically runs tests for packages that have been modified.

### How it works

1. Detects which Go files are staged for commit
2. Identifies affected packages (auth, product, order)
3. Runs tests only for affected packages
4. Blocks commit if any tests fail

### Installation

Run from project root:

```bash
make install-hooks
```

Or manually:

```bash
chmod +x scripts/hooks/pre-commit.sh
cp scripts/hooks/pre-commit.sh .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit
```

### Usage

Once installed, the hook runs automatically when you commit:

```bash
git add internal/auth/jwt.go
git commit -m "Update JWT implementation"
```

If tests fail, the commit will be aborted with an error message.

### Bypassing the hook

In emergency situations, you can skip the hook with:

```bash
git commit --no-verify -m "Emergency fix"
```

Use this sparingly as it bypasses test validation.

### Testing the hook manually

```bash
make pre-commit
```

Or directly:

```bash
.git/hooks/pre-commit
```

## Supported Packages

The hook currently tests these packages when changes are detected:

- `internal/auth` - Authentication and user management
- `internal/product` - Product management
- `internal/order` - Order management

To add more packages, edit `scripts/hooks/pre-commit.sh` and add matching conditions.
