# Go 项目 Makefile

# 定义变量
BINARY_NAME=doghole
PKG_PATH=doghole/cmd
VERSION_VAR_PATH=$(PKG_PATH).Version
COMMIT_HASH_VAR_PATH=$(PKG_PATH).CommitHash
BUILD_TIME_VAR_PATH=$(PKG_PATH).BuildTime

# Go 命令
GO=go
GO_BUILD=$(GO) build
GO_CLEAN=$(GO) clean
GO_RUN=$(GO) run
GO_TEST=$(GO) test
GO_LINT=golangci-lint run # 假设您使用 golangci-lint

# 获取 Git commit hash 和构建时间
GIT_COMMIT_HASH=$(shell git rev-parse --short HEAD)
BUILD_TIME=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')

# 构建标志，用于注入版本信息
LDFLAGS_VERSION = -X "$(VERSION_VAR_PATH)=$(shell git describe --tags --always --dirty)" \
                  -X "$(COMMIT_HASH_VAR_PATH)=$(GIT_COMMIT_HASH)" \
                  -X "$(BUILD_TIME_VAR_PATH)=$(BUILD_TIME)"

# 默认目标
all: build

# 构建应用程序
build:
	@echo "Building $(BINARY_NAME)..."
	@$(GO_BUILD) -ldflags="$(LDFLAGS_VERSION)" -o $(BINARY_NAME) main.go
	@echo "Build complete: $(BINARY_NAME)"

# 运行应用程序 (开发模式)
run:
	@echo "Running $(BINARY_NAME) (dev mode)..."
	@$(GO_RUN) main.go server --config config.yaml

# 清理构建产物
clean:
	@echo "Cleaning..."
	@$(GO_CLEAN)
	@rm -f $(BINARY_NAME)
	@echo "Clean complete."

# 运行测试
test:
	@echo "Running tests..."
	@$(GO_TEST) ./...

# 运行 linter (需要安装 golangci-lint)
lint:
	@echo "Running linter..."
	@$(GO_LINT) ./...

# 显示帮助信息
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  all          Build the application (default)."
	@echo "  build        Build the application with version information."
	@echo "  run          Run the application in development mode."
	@echo "  clean        Remove build artifacts."
	@echo "  test         Run tests."
	@echo "  lint         Run linter (requires golangci-lint)."
	@echo "  help         Show this help message."

.PHONY: all build run clean test lint help
