.PHONY: migrate-up migrate-down migrate-create migrate-force migrate-version setup install-hooks test test-auth test-product test-order pre-commit

migrate-up:
	@echo "Running migrations..."
	migrate -path migrations -database "${DATABASE_URL}" up

migrate-down:
	@echo "Rolling back last migration..."
	migrate -path migrations -database "${DATABASE_URL}" down 1

migrate-create:
	@if [ -z "$(name)" ]; then \
		echo "Error: Please provide a migration name using name=your_migration_name"; \
		exit 1; \
	fi
	@echo "Creating migration: $(name)"
	migrate create -ext sql -dir migrations -seq $(name)

migrate-force:
	@if [ -z "$(version)" ]; then \
		echo "Error: Please provide a version using version=N"; \
		exit 1; \
	fi
	@echo "Forcing migration version to $(version)..."
	migrate -path migrations -database "${DATABASE_URL}" force $(version)

migrate-version:
	@echo "Current migration version:"
	migrate -path migrations -database "${DATABASE_URL}" version

setup:
	@echo "Setting up development environment..."
	@chmod +x scripts/setup.sh
	@./scripts/setup.sh
	@echo "Setup complete! You can now start developing."

install-hooks:
	@echo "Configuring git hooks..."
	@git config core.hooksPath .githooks
	@chmod +x .githooks/pre-commit
	@echo "Git hooks configured successfully"

test:
	@go test ./... -v

test-auth:
	@go test ./internal/auth -v

test-product:
	@go test ./internal/product -v

test-order:
	@go test ./internal/order -v

pre-commit:
	@.githooks/pre-commit
