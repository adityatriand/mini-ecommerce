# Git Hooks for Mini E-Commerce Microservices

This directory contains Git hooks for the Mini E-Commerce microservices project.

## Pre-commit Hook

The pre-commit hook automatically runs tests and code quality checks before each commit.

### What it does:

1. **Service Detection** - Identifies which microservices are affected by the changes
2. **Code Quality Checks**:
   - `go mod tidy` - Ensures dependencies are clean
   - `go vet` - Static analysis
   - `go fmt` - Code formatting check
3. **Testing**:
   - Unit tests for affected services
   - Test coverage reporting
   - Shared library tests (if shared code is modified)
4. **Integration Tests** - Runs integration tests if available
5. **TODO/FIXME Detection** - Warns about pending items

### Installation:

The pre-commit hook is automatically installed when you run the setup script:

```bash
./scripts/setup-microservices.sh
```

Or manually:

```bash
cp .githooks/pre-commit .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit
```

### Usage:

The hook runs automatically on every commit. If tests fail, the commit will be blocked until issues are resolved.

### Bypassing the hook:

If you need to bypass the hook (not recommended), use:

```bash
git commit --no-verify -m "your message"
```

## Customization

You can modify the pre-commit hook to add additional checks:

- Security scanning
- Performance tests
- Documentation checks
- Custom linting rules

## Troubleshooting

### Common Issues:

1. **Hook not running**: Ensure the file is executable (`chmod +x .git/hooks/pre-commit`)
2. **Tests failing**: Fix the failing tests before committing
3. **Formatting issues**: Run `go fmt ./...` in the affected service directory
4. **Dependency issues**: Run `go mod tidy` in the affected service directory

### Debug Mode:

To see more detailed output, you can modify the hook to add `set -x` at the beginning for debug mode.