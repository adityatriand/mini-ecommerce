# Automated Setup for Contributors

This project uses automated git hooks to ensure code quality. Unlike projects that require manual setup, hooks are automatically installed when contributors run the setup command.

## How It Works

When a new contributor clones the repository and runs:

```bash
make setup
```

The following happens automatically:
1. Git hooks are installed to `.git/hooks/`
2. Go dependencies are downloaded
3. Development tools are installed
4. Pre-commit testing is enabled

## What This Means for Contributors

Contributors do NOT need to:
- Manually install hooks
- Remember to run tests before committing
- Install additional tools

The setup is one command and automatic.

## Pre-Commit Hook Behavior

Once installed, every commit automatically:
1. Detects which Go files were changed
2. Identifies affected packages (auth, product, order)
3. Runs tests only for those packages
4. Blocks commit if tests fail

## Comparison to Manual Setup

### Without Automation (Manual)
```bash
git clone <repo>
cd mini-ecommerce
go mod download
chmod +x .git/hooks/pre-commit
cp scripts/hooks/pre-commit.sh .git/hooks/pre-commit
```

### With Automation (Current)
```bash
git clone <repo>
cd mini-ecommerce
make setup
```

## Benefits

1. Consistent setup across all developers
2. No forgotten steps
3. Hooks always up-to-date
4. New contributors can start immediately
5. Similar experience to Husky in Node.js projects

## For Project Maintainers

To update hooks for all contributors:
1. Update `scripts/hooks/pre-commit.sh`
2. Contributors run `make setup` to get latest version
3. Changes are version controlled in git

## Troubleshooting

If hooks are not working:
```bash
make install-hooks
```

To verify hooks are installed:
```bash
ls -la .git/hooks/pre-commit
```

Should show an executable file.
