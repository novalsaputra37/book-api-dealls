APP_NAME=github.com/adf-code/beta-book-api
BUILD_DIR=bin
CMD_ENTRY=cmd/main.go
SWAG=swag

.PHONY: all swag build run dev clean consumer

all: dev

# Install Deps
install:
	@echo "🧩 Installing dependency packages..."
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/vektra/mockery/v2@latest

# Generate Swagger docs
unit-test:
	@echo "🧲 Starting unit test..."
	go test ./internal/usecase -v

# Generate Swagger docs
swag:
	@echo "📚 Generating Swagger docs..."
	$(SWAG) init -g $(CMD_ENTRY) -o ./docs

# Build binary
build:
	@echo "🔨 Building app binary..."
	go build -o $(BUILD_DIR)/$(APP_NAME) $(CMD_ENTRY)

# Run binary
run:
	@echo "🚀 Running app..."
	./$(BUILD_DIR)/$(APP_NAME)

# Dev: Generate Swagger + Build + Run
dev:
	@$(MAKE) swag
	@$(MAKE) build
	@$(MAKE) run

# Docker Build and Run
docker-build-run:
	docker-compose up -d

# Docker Rebuild and Run
docker-rebuild-run:
	docker-compose down --remove-orphans --volumes
	docker-compose build --no-cache
	docker-compose up -d

# Clean build
clean:
	@echo "🧹 Cleaning build directory..."
	rm -rf $(BUILD_DIR)

# Build consumer binary
build-consumer:
	@echo "🔨 Building consumer binary..."
	go build -o $(BUILD_DIR)/$(APP_NAME)-consumer cmd/consumer/main.go

# Run consumer
run-consumer:
	@echo "🚀 Running consumer..."
	./$(BUILD_DIR)/$(APP_NAME)-consumer

# Dev consumer: Build + Run
consumer:
	@$(MAKE) build-consumer
	@$(MAKE) run-consumer
