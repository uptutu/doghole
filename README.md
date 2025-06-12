# Doghole

My App Server Skeleton. Build with fiber\ent\zap.

[![Go Report Card](https://goreportcard.com/badge/github.com/uptutu/doghole)](https://goreportcard.com/report/github.com/uptutu/doghole)
[![Go Version](https://img.shields.io/github/go-mod/go-version/uptutu/doghole)](https://golang.org/)
[![License](https://img.shields.io/github/license/uptutu/doghole)](LICENSE)
[![Build Status](https://img.shields.io/github/actions/workflow/status/uptutu/doghole/go.yml?branch=main)](https://github.com/uptutu/doghole/actions)

**Doghole** 是一个使用 Go 语言编写的现代化、高性能的 Web 应用程序骨架。它集成了以下优秀的技术栈：

- **Fiber**: 一个受 Express.js 启发的 Go Web 框架，以高性能和低内存占用著称。
- **Ent**: 一个简单而强大的 Go 实体框架，用于轻松管理数据库模式和查询。
- **Zap**: 一个极快、结构化的 Go 日志库。
- **Cobra**: 一个强大的 Go CLI 应用程序库。
- **Viper**: 一个完整的 Go 应用程序配置解决方案。

## ✨ 特性

- **模块化设计**: 清晰的项目结构，易于扩展和维护。
- **配置驱动**: 通过 YAML 文件或环境变量轻松配置应用程序。
- **结构化日志**: 使用 Zap 实现高性能的结构化日志记录。
- **数据库集成**: 使用 Ent 进行类型安全的数据库操作和迁移。
- **CLI 支持**: 使用 Cobra 构建强大的命令行界面。
- **优雅关闭**: 服务器支持优雅关闭，确保请求处理完成。
- **API 版本控制**: 内置 API 版本控制机制。
- **中间件支持**: 易于添加和管理中间件（如 CORS、Logger、RequestID）。
- **Makefile 支持**: 包含一个 Makefile，用于简化构建、运行、测试和清理任务。
- **版本信息注入**: 构建时自动注入版本号、Git Commit Hash 和构建时间。

## 🚀 快速开始

### 先决条件

- [Go](https://golang.org/dl/) (版本 1.20 或更高)
- [Docker](https://www.docker.com/get-started) (可选, 用于数据库)
- [Make](https://www.gnu.org/software/make/) (用于执行 Makefile 命令)

### 安装

1.  **克隆仓库**:
    ```bash
    git clone https://github.com/your_username/doghole.git
    cd doghole
    ```

2.  **安装依赖**:
    ```bash
    go mod tidy
    ```

### 配置

1.  复制示例配置文件:
    ```bash
    cp config.example.yaml config.yaml
    ```
2.  根据您的环境修改 `config.yaml` 文件，特别是数据库连接信息。

### 运行

有多种方式可以运行此应用程序：

1.  **使用 `go run` (开发模式)**:
    ```bash
    go run main.go server --config config.yaml
    ```

2.  **使用 `make run` (推荐的开发模式)**:
    ```bash
    make run
    ```

3.  **构建并运行二进制文件**:
    ```bash
    make build
    ./doghole server --config config.yaml
    ```

服务器默认启动在 `http://localhost:8080`。

## 🛠️ Makefile 命令

项目包含一个 `Makefile` 来简化常见的开发任务：

-   `make build`: 构建应用程序二进制文件。
-   `make run`: 在开发模式下运行应用程序。
-   `make clean`: 清理构建产物。
-   `make test`: 运行单元测试。
-   `make lint`: 运行 Go linter (需要安装 `golangci-lint`)。
-   `make help`: 显示所有可用的 Makefile 命令。

## 📦 构建

要构建生产环境的二进制文件，包含版本信息：

```bash
make build
```

这将生成一个名为 `doghole` (或您在 Makefile 中配置的名称) 的可执行文件。

## 📝 版本控制

应用程序支持在构建时注入版本信息。您可以通过以下命令查看版本：

```bash
./doghole version
```

或者，如果您正在运行服务，通常会有一个 `/version` 或类似的管理端点来显示此信息。

## ⚙️ 配置选项

应用程序可以通过 `config.yaml` 文件或环境变量进行配置。详细的配置选项请参考 `config/config.go` 文件中的结构体定义。

主要配置部分包括：

-   `server`: HTTP 服务器配置 (端口、超时等)
-   `db`: 数据库连接配置 (支持主从库)
-   `logger`: 日志系统配置 (级别、格式、输出等)

## 🤝 贡献

欢迎贡献！如果您想为 Doghole 做出贡献，请遵循以下步骤：

1.  Fork 本仓库。
2.  创建一个新的分支 (`git checkout -b feature/your-feature-name`)。
3.  提交您的更改 (`git commit -am 'Add some feature'`)。
4.  将您的分支推送到远程仓库 (`git push origin feature/your-feature-name`)。
5.  创建一个新的 Pull Request。

请确保您的代码符合项目编码规范，并通过所有测试。

## 📄 许可证

本项目采用 [MIT 许可证](LICENSE)。
