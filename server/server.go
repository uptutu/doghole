package server

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"go.uber.org/zap"
)

// ServerConfig 定义服务器配置项
type ServerConfig struct {
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ShutdownTimeout   time.Duration
	EnableCompression bool
	EnablePrefork     bool
}

// DefaultConfig 返回默认服务器配置
func DefaultConfig() ServerConfig {
	return ServerConfig{
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
		ShutdownTimeout:   5 * time.Second,
		EnableCompression: true,
		EnablePrefork:     false,
	}
}

// Server 表示HTTP服务器
type Server struct {
	app    *fiber.App
	config ServerConfig
	logger *zap.Logger
}

// NewServer 创建一个新的服务器实例
func NewServer(options ...func(*Server)) *Server {
	// 创建带有默认配置的服务器
	s := &Server{
		config: DefaultConfig(),
		logger: zap.L(),
	}

	// 应用所有选项
	for _, option := range options {
		option(s)
	}

	// 配置Fiber应用
	fiberConfig := fiber.Config{
		ReadTimeout:          s.config.ReadTimeout,
		WriteTimeout:         s.config.WriteTimeout,
		IdleTimeout:          s.config.IdleTimeout,
		CompressedFileSuffix: ".gz",
		EnableCompression:    s.config.EnableCompression,
		Prefork:              s.config.EnablePrefork,
	}

	s.app = fiber.New(fiberConfig)

	// 添加全局中间件
	s.app.Use(recover.New())

	// 注册路由
	RegisterRoutes(s.app)

	return s
}

// WithConfig 设置服务器配置
func WithConfig(config ServerConfig) func(*Server) {
	return func(s *Server) {
		s.config = config
	}
}

// WithLogger 设置日志记录器
func WithLogger(logger *zap.Logger) func(*Server) {
	return func(s *Server) {
		s.logger = logger
	}
}

// Start 启动HTTP服务器
func (s *Server) Start(addr string) error {
	// 设置优雅关闭
	go s.gracefulShutdown()

	s.logger.Info("服务器启动", zap.String("地址", addr))
	return s.app.Listen(addr)
}

// gracefulShutdown 优雅关闭服务器
func (s *Server) gracefulShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit
	s.logger.Info("正在关闭服务器...")

	ctx, cancel := context.WithTimeout(context.Background(), s.config.ShutdownTimeout)
	defer cancel()

	if err := s.app.ShutdownWithContext(ctx); err != nil {
		s.logger.Fatal("服务器强制关闭", zap.Error(err))
	}

	s.logger.Info("服务器已优雅关闭")
}

// App 返回底层的Fiber应用实例
func (s *Server) App() *fiber.App {
	return s.app
}
