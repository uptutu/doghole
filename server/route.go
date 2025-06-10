package server

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"go.uber.org/zap"
)

// Router 路由器接口
type Router interface {
	Setup(app *fiber.App)
}

// CustomLogger 自定义日志中间件
func CustomLogger() fiber.Handler {
	return func(ctx fiber.Ctx) error {
		start := time.Now()
		requestID := ctx.Locals("requestid")

		// 记录请求信息
		zap.L().Debug("请求接收",
			zap.String("method", ctx.Method()),
			zap.String("path", ctx.Path()),
			zap.String("ip", ctx.IP()),
			zap.Any("requestID", requestID),
			zap.Any("headers", ctx.GetReqHeaders()),
		)

		// 处理请求
		err := ctx.Next()

		// 记录响应信息
		latency := time.Since(start)
		status := ctx.Response().StatusCode()

		logFunc := zap.L().Debug
		if status >= 400 {
			logFunc = zap.L().Warn
		}
		if status >= 500 {
			logFunc = zap.L().Error
		}

		logFunc("响应发送",
			zap.Int("status", status),
			zap.Duration("latency", latency),
			zap.String("method", ctx.Method()),
			zap.String("path", ctx.Path()),
			zap.Any("requestID", requestID),
		)

		return err
	}
}

// RegisterRoutes 注册所有路由和中间件
func RegisterRoutes(app *fiber.App) {
	// 全局中间件
	app.Use(
		requestid.New(), // 请求ID中间件
		logger.New(logger.Config{ // Fiber内置日志中间件
			Format:     "${pid} ${locals:requestid} ${status} - ${method} ${path}\n",
			TimeFormat: "2006-01-02 15:04:05",
			TimeZone:   "Asia/Shanghai",
		}),
		CustomLogger(), // 自定义日志中间件
		cors.New(cors.Config{ // CORS中间件
			AllowOrigins:     "*",
			AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
			AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
			AllowCredentials: true,
			MaxAge:           300,
		}),
	)

	// API版本控制
	api := app.Group("/api")
	v1 := api.Group("/v1")

	// 健康检查路由
	app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// 根路由
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("Welcome to Doghole API")
	})

	// 注册API路由
	registerV1Routes(v1)
}

// registerV1Routes 注册V1版本的API路由
func registerV1Routes(router fiber.Router) {
	// 用户相关路由
	userGroup := router.Group("/users")
	userGroup.Get("/", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "获取所有用户"})
	})
	userGroup.Post("/", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "创建用户"})
	})
	userGroup.Get("/:id", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "获取单个用户", "id": c.Params("id")})
	})
	userGroup.Put("/:id", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "更新用户", "id": c.Params("id")})
	})
	userGroup.Delete("/:id", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "删除用户", "id": c.Params("id")})
	})

	// 这里可以继续添加其他路由组
}
