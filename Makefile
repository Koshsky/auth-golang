# Переменные окружения
export PATH := $(HOME)/go/bin:$(PATH)
export GOPATH := $(HOME)/go

.PHONY: proto generate test build clean

# Генерация Go кода из proto файлов
proto:
	@echo "Generating Go code from proto files..."
	cd api/authpb && protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative auth.proto
	@echo "Proto generation completed!"

# Генерация моков
generate:
	@echo "Generating mocks..."
	$(shell which go) generate ./...
	@echo "Mock generation completed!"

# Запуск тестов
test:
	@echo "Running tests with coverage..."
	@mkdir -p coverage
	$(shell which go) test -coverprofile=coverage/coverage.out ./...
	$(shell which go) tool cover -html=coverage/coverage.out -o coverage/coverage.html
	@echo "Generating package-specific coverage reports..."
	@for pkg in $$(go list ./... | grep -v /vendor/); do \
		pkg_path=$$(echo $$pkg | sed 's|github.com/Koshsky/subs-service/auth-service/||'); \
		pkg_dir=$$(dirname coverage/$$pkg_path); \
		pkg_name=$$(basename $$pkg_path); \
		echo "Testing $$pkg_path..."; \
		mkdir -p $$pkg_dir; \
		go test -coverprofile=$$pkg_dir/$$pkg_name.out $$pkg; \
		if [ -f $$pkg_dir/$$pkg_name.out ] && [ -s $$pkg_dir/$$pkg_name.out ] && [ $$(wc -l < $$pkg_dir/$$pkg_name.out) -gt 1 ]; then \
			# Проверяем, есть ли реальное покрытие (не все строки с покрытием 0) \
			if grep -q " [1-9]$$" $$pkg_dir/$$pkg_name.out; then \
				go tool cover -html=$$pkg_dir/$$pkg_name.out -o $$pkg_dir/$$pkg_name.html; \
				echo "Generated coverage report: $$pkg_dir/$$pkg_name.html"; \
			else \
				echo "No test coverage for $$pkg_path (all lines uncovered), skipping HTML generation"; \
			fi; \
		else \
			echo "No tests found for $$pkg_path, skipping HTML generation"; \
		fi; \
		# Удаляем .out файл в любом случае \
		rm -f $$pkg_dir/$$pkg_name.out; \
	done
	@echo "Cleaning up empty directories..."
	find coverage -type d -empty -delete
	@echo "Tests completed! Coverage reports saved in coverage/ directory"

# Запуск линтера
lint:
	@echo "Running linter..."
	golangci-lint run --verbose
	@echo "Linting completed!"

# Сборка приложения
build:
	@echo "Building application..."
	$(shell which go) build -o bin/auth-service ./cmd/auth-service
	@echo "Build completed!"

# Очистка
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -rf coverage/
	@echo "Removing all coverage files recursively..."
	find . -name "*.out" -type f -delete
	find . -name "*.html" -path "*/coverage*" -type f -delete
	@echo "Clean completed!"

# Полная сборка (proto + generate + build)
all: proto generate build

# Помощь
help:
	@echo "Available targets:"
	@echo "  proto     - Generate Go code from proto files"
	@echo "  generate  - Generate mocks"
	@echo "  test      - Run tests"
	@echo "  lint      - Run linter"
	@echo "  build     - Build application"
	@echo "  clean     - Clean build artifacts"
	@echo "  all       - Run proto, generate, and build"
	@echo "  help      - Show this help message"
