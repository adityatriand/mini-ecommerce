# Contributing Guide

Thank you for contributing to this project!

## Getting Started

### First Time Setup

After cloning the repository, run:

```bash
make setup
```

This configures:
- Git hooks path to use `.githooks` directory (version controlled)
- Downloads Go dependencies
- Installs development tools

After setup, hooks work automatically because they are committed to the repository.

### What Gets Configured

The setup configures a pre-commit hook that:
- Runs tests for packages you modified
- Prevents commits if tests fail
- Only tests affected code (auth, product, or order)

## Development Workflow

### Making Changes

1. Create a new branch:
```bash
git checkout -b feat/your-feature
```

2. Make your changes

3. Write tests for your changes

4. Commit your changes:
```bash
git add .
git commit -m "Add your feature"
```

The pre-commit hook will automatically run tests for affected packages.

### Running Tests

Run all tests:
```bash
make test
```

Run specific package tests:
```bash
make test-auth
make test-product
make test-order
```

Test the pre-commit hook manually:
```bash
make pre-commit
```

### Commit Guidelines

The pre-commit hook ensures:
- All tests pass before commit
- Only affected packages are tested
- Fast feedback loop

If tests fail, fix them before committing.

### Bypass Hook (Emergency Only)

In rare cases, you can skip the hook:
```bash
git commit --no-verify -m "Emergency fix"
```

Use this sparingly.

## Project Structure

```
.
├── cmd/                  # Application entrypoints
├── internal/
│   ├── auth/            # Authentication package
│   ├── product/         # Product management
│   └── order/           # Order management
├── scripts/
│   ├── hooks/           # Git hooks
│   └── setup.sh         # Setup script
└── Makefile             # Build commands
```

## Writing Tests

### Test File Naming

Tests should be in the same package as the code being tested:
```
internal/auth/service.go
internal/auth/service_test.go
```

### Test Coverage

When adding new features:
1. Write unit tests for all public functions
2. Test both success and error cases
3. Use mocks for external dependencies

### Example Test

```go
func TestService_RegisterUser(t *testing.T) {
    t.Run("should register user successfully", func(t *testing.T) {
        mockRepo := new(MockRepository)
        service := NewService(mockRepo, ...)

        input := RegisterRequest{
            Email:    "test@example.com",
            Password: "password123",
        }

        mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

        user, err := service.RegisterUser(ctx, input)

        require.NoError(t, err)
        assert.NotNil(t, user)
        mockRepo.AssertExpectations(t)
    })
}
```

## Questions?

If you have questions or run into issues, please open an issue on GitHub.
