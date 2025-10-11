# Git Hooks

This directory contains git hooks that are version controlled and shared across all developers.

## How It Works

Unlike standard git hooks in `.git/hooks/` which cannot be committed, this project uses Git's `core.hooksPath` configuration to point to this `.githooks` directory.

When you run `make setup`, it configures your local git to use hooks from this directory:

```bash
git config core.hooksPath .githooks
```

This means:
1. Hooks are version controlled
2. All developers get the same hooks
3. Hook updates are automatically pulled with git pull
4. One-time setup per developer

## Pre-commit Hook

Automatically runs tests for modified packages before allowing commits.

### What It Does

1. Detects staged Go files
2. Identifies affected packages (auth, product, order)
3. Runs tests only for those packages
4. Blocks commit if tests fail

### Setup

First time after cloning:
```bash
make setup
```

This configures git to use hooks from this directory.

### Manual Configuration

If you prefer manual setup:
```bash
git config core.hooksPath .githooks
```

### Bypass Hook

In emergencies only:
```bash
git commit --no-verify -m "Emergency fix"
```

## Benefits Over Traditional Hooks

Traditional `.git/hooks/`:
- Not version controlled
- Manual installation required
- Inconsistent across team
- Difficult to update

This `.githooks/` approach:
- Version controlled in git
- Automatic with one-time setup
- Consistent for everyone
- Updates via git pull
