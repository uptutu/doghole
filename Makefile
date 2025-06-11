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

# Docker 命令
DOCKER=docker
DOCKER_BUILD=$(DOCKER) build
DOCKER_IMAGE_NAME=$(BINARY_NAME)
# 使用 Git commit hash 作为默认的 Docker 镜像标签
LATEST_COMMIT_HASH=$(shell git rev-parse --short HEAD)
DOCKER_TAG=$(LATEST_COMMIT_HASH)
APP_VERSION_TAG=$(shell git describe --tags --always --dirty)

# 获取 Git commit hash 和构建时间
GIT_COMMIT_HASH=$(shell git rev-parse --short HEAD)
BUILD_TIME=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')

# 构建标志，用于注入版本信息
LDFLAGS_VERSION = -X "$(VERSION_VAR_PATH)=$(APP_VERSION_TAG)" \
                  -X "$(COMMIT_HASH_VAR_PATH)=$(GIT_COMMIT_HASH)" \
                  -X "$(BUILD_TIME_VAR_PATH)=$(BUILD_TIME)"

# 默认目标
all: build

# 构建应用程序
build:
	@echo "Building $(BINARY_NAME)..."
	@$(GO_BUILD) -ldflags="$(LDFLAGS_VERSION)" -o $(BINARY_NAME) main.go
	@echo "Build complete: $(BINARY_NAME)"

# 构建 Docker 镜像
docker-build:
	@echo "Building Docker image $(DOCKER_IMAGE_NAME):$(DOCKER_TAG)..."
	@$(DOCKER_BUILD) --build-arg APP_VERSION=$(APP_VERSION_TAG) \
	                 --build-arg COMMIT_HASH=$(GIT_COMMIT_HASH) \
	                 --build-arg BUILD_TIME=$(BUILD_TIME) \
	                 -t $(DOCKER_IMAGE_NAME):$(DOCKER_TAG) \
	                 -t $(DOCKER_IMAGE_NAME):latest .
	@echo "Docker image $(DOCKER_IMAGE_NAME):$(DOCKER_TAG) built successfully."
	@echo "Also tagged as $(DOCKER_IMAGE_NAME):latest."

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
	@echo "  docker-build Build the Docker image for the application."
	@echo "  help         Show this help message."

.PHONY: all build run clean test lint docker-build help
