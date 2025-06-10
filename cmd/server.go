package cmd

import (
	"context"
	"fmt"
	"os"

	"doghole/config"
	"doghole/domain/conn"
	"doghole/ent"
	"doghole/logger"
	"doghole/server"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var _config *string

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "启动HTTP服务器",
	Long:  `此命令启动Doghole HTTP服务器，该服务器侦听传入请求并根据配置的设置处理它们。`,
	Run: func(cmd *cobra.Command, args []string) {
		// 加载配置文件
		conf := config.NewConfig()
		config.SetGlobalConfig(conf)

		if err := conf.LoadSingleConfigFile(*_config); err != nil {
			fmt.Printf("加载配置文件失败: %s\n", err)
			os.Exit(1)
		}

		// 初始化日志系统
		err := logger.Init(
			logger.WithLevel(conf.Logger.Level),
			logger.WithFormat(conf.Logger.Format),
			logger.WithOutputPath(conf.Logger.Outfile),
			logger.WithRotation(conf.Logger.ChuckSize, 3, 7, true),
			logger.WithDevelopment(conf.Logger.Level == "debug"),
			logger.WithField("app", "doghole"),
		)
		if err != nil {
			fmt.Printf("初始化日志系统失败: %s\n", err)
			os.Exit(1)
		}

		// 确保在程序退出时正确关闭资源
		defer func() {
			conn.Close()
			logger.Sync()
		}()

		// 初始化数据库连接
		ctx := context.Background()

		if conf.DB.DB != nil {
			// 单一数据库配置
			if conf.DB.DB.ToDialect() == "" || conf.DB.DB.ToDNS() == "" {
				zap.L().Fatal("数据库配置错误", zap.String("error", "数据库配置不能为空"))
			}

			db, err := ent.Open(conf.DB.DB.ToDialect(), conf.DB.DB.ToDNS())
			if err != nil {
				zap.L().Fatal("连接数据库失败", zap.Error(err))
			}

			// 设置共享连接
			if err := conn.Initialize(ctx, db, db); err != nil {
				zap.L().Fatal("初始化数据库连接失败", zap.Error(err))
			}
		} else {
			// 读写分离配置
			if conf.DB.WriteDB == nil || conf.DB.ReadDB == nil {
				zap.L().Fatal("数据库配置错误", zap.String("error", "必须至少提供一个读或写数据库配置"))
			}

			// 初始化读取连接
			reader, err := ent.Open(conf.DB.ReadDB.ToDialect(), conf.DB.ReadDB.ToDNS())
			if err != nil {
				zap.L().Fatal("连接读取数据库失败", zap.Error(err))
			}

			// 初始化写入连接
			writer, err := ent.Open(conf.DB.WriteDB.ToDialect(), conf.DB.WriteDB.ToDNS())
			if err != nil {
				zap.L().Fatal("连接写入数据库失败", zap.Error(err))
			}

			// 初始化连接管理器
			if err := conn.Initialize(ctx, writer, reader); err != nil {
				zap.L().Fatal("初始化数据库连接失败", zap.Error(err))
			}
		}

		// 创建数据库架构
		if err := conn.Writer().Schema.Create(ctx); err != nil {
			zap.L().Fatal("创建数据库架构失败", zap.Error(err))
		}

		// 创建服务器
		serverConfig := server.ServerConfig{
			ReadTimeout:       conf.Server.ReadTimeout,
			WriteTimeout:      conf.Server.WriteTimeout,
			IdleTimeout:       conf.Server.IdleTimeout,
			ShutdownTimeout:   conf.Server.ShutdownTimeout,
			EnableCompression: conf.Server.EnableCompression,
			EnablePrefork:     conf.Server.EnablePrefork,
		}

		// 使用选项模式创建服务器
		srv := server.NewServer(
			server.WithConfig(serverConfig),
			server.WithLogger(zap.L()),
		)

		// 启动服务器
		zap.L().Info("服务器已启动", zap.Int("port", conf.Server.Port))
		if err := srv.Start(conf.ToPort()); err != nil {
			zap.L().Fatal("服务器启动失败", zap.Error(err))
		}
	},
}

func init() {
	// 定义命令行标志
	_config = serverCmd.Flags().StringP("config", "c", "config.yaml", "配置文件路径")
	serverCmd.MarkFlagRequired("config") // 将config标志设为必需

	// 添加其他可选标志
	serverCmd.Flags().IntP("port", "p", 0, "服务器端口（覆盖配置文件）")
	serverCmd.Flags().StringP("log-level", "l", "", "日志级别（覆盖配置文件）")

	// 添加服务器命令到根命令
	rootCmd.AddCommand(serverCmd)
}
