package logger

import (
	"io"
	"os"
	"strings"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// 日志配置常量
const (
	// 日志级别
	LevelDebug = "debug"
	LevelInfo  = "info"
	LevelWarn  = "warn"
	LevelError = "error"
	LevelFatal = "fatal"

	// 日志格式
	FormatJSON    = "json"
	FormatConsole = "console"
)

// LoggerOptions 日志选项
type LoggerOptions struct {
	Level           string            // 日志级别
	Format          string            // 日志格式
	OutputPath      string            // 输出文件路径
	ErrorOutputPath string            // 错误输出路径
	MaxSize         int               // 单个日志文件最大大小(MB)
	MaxBackups      int               // 最大保留旧日志文件数
	MaxAge          int               // 日志文件保留天数
	Compress        bool              // 是否压缩日志文件
	AddCaller       bool              // 是否添加调用者信息
	Development     bool              // 是否为开发模式
	Fields          map[string]string // 全局日志字段
}

// DefaultLoggerOptions 返回默认日志选项
func DefaultLoggerOptions() LoggerOptions {
	return LoggerOptions{
		Level:           LevelInfo,
		Format:          FormatJSON,
		OutputPath:      "stdout",
		ErrorOutputPath: "stderr",
		MaxSize:         100,
		MaxBackups:      3,
		MaxAge:          7,
		Compress:        true,
		AddCaller:       true,
		Development:     false,
		Fields:          map[string]string{},
	}
}

// Option 日志选项函数
type Option func(*LoggerOptions)

// WithLevel 设置日志级别
func WithLevel(level string) Option {
	return func(o *LoggerOptions) {
		o.Level = level
	}
}

// WithFormat 设置日志格式
func WithFormat(format string) Option {
	return func(o *LoggerOptions) {
		o.Format = format
	}
}

// WithOutputPath 设置输出路径
func WithOutputPath(path string) Option {
	return func(o *LoggerOptions) {
		o.OutputPath = path
	}
}

// WithErrorOutputPath 设置错误输出路径
func WithErrorOutputPath(path string) Option {
	return func(o *LoggerOptions) {
		o.ErrorOutputPath = path
	}
}

// WithRotation 设置日志轮转参数
func WithRotation(maxSize, maxBackups, maxAge int, compress bool) Option {
	return func(o *LoggerOptions) {
		o.MaxSize = maxSize
		o.MaxBackups = maxBackups
		o.MaxAge = maxAge
		o.Compress = compress
	}
}

// WithCaller 设置是否添加调用者信息
func WithCaller(addCaller bool) Option {
	return func(o *LoggerOptions) {
		o.AddCaller = addCaller
	}
}

// WithDevelopment 设置是否为开发模式
func WithDevelopment(development bool) Option {
	return func(o *LoggerOptions) {
		o.Development = development
	}
}

// WithField 添加全局日志字段
func WithField(key, value string) Option {
	return func(o *LoggerOptions) {
		o.Fields[key] = value
	}
}

// WithOutfile 兼容旧API，设置输出文件路径和大小
func WithOutfile(outfile string, chuckSize int) Option {
	return func(o *LoggerOptions) {
		if outfile != "" {
			o.OutputPath = outfile
			o.MaxSize = chuckSize
		}
	}
}

// getLogLevel 获取zapcore日志级别
func getLogLevel(level string) zapcore.Level {
	switch strings.ToLower(level) {
	case LevelDebug:
		return zapcore.DebugLevel
	case LevelInfo:
		return zapcore.InfoLevel
	case LevelWarn:
		return zapcore.WarnLevel
	case LevelError:
		return zapcore.ErrorLevel
	case LevelFatal:
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// Init 初始化日志系统
func Init(opts ...Option) error {
	options := DefaultLoggerOptions()

	// 应用选项
	for _, opt := range opts {
		opt(&options)
	}

	// 配置编码器
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 开发模式下使用更友好的编码器
	if options.Development {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoderConfig.EncodeCaller = zapcore.FullCallerEncoder
	}

	// 配置编码器
	var encoder zapcore.Encoder
	if options.Format == FormatJSON {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// 配置输出
	var output zapcore.WriteSyncer
	if options.OutputPath == "stdout" {
		output = zapcore.AddSync(os.Stdout)
	} else if options.OutputPath == "stderr" {
		output = zapcore.AddSync(os.Stderr)
	} else {
		// 文件输出，使用lumberjack进行日志轮转
		ljLogger := &lumberjack.Logger{
			Filename:   options.OutputPath,
			MaxSize:    options.MaxSize,
			MaxBackups: options.MaxBackups,
			MaxAge:     options.MaxAge,
			Compress:   options.Compress,
		}
		output = zapcore.AddSync(ljLogger)
	}

	// 配置错误输出
	var errorOutput zapcore.WriteSyncer
	if options.ErrorOutputPath == "stderr" {
		errorOutput = zapcore.AddSync(os.Stderr)
	} else if options.ErrorOutputPath == "stdout" {
		errorOutput = zapcore.AddSync(os.Stdout)
	} else {
		// 文件输出，使用lumberjack进行日志轮转
		ljLogger := &lumberjack.Logger{
			Filename:   options.ErrorOutputPath,
			MaxSize:    options.MaxSize,
			MaxBackups: options.MaxBackups,
			MaxAge:     options.MaxAge,
			Compress:   options.Compress,
		}
		errorOutput = zapcore.AddSync(ljLogger)
	}

	// 创建Core
	core := zapcore.NewCore(
		encoder,
		output,
		zap.NewAtomicLevelAt(getLogLevel(options.Level)),
	)

	// 添加全局字段
	fields := make([]zap.Field, 0, len(options.Fields))
	for k, v := range options.Fields {
		fields = append(fields, zap.String(k, v))
	}

	// 创建Logger
	logger := zap.New(core,
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
		zap.ErrorOutput(errorOutput),
		zap.Fields(fields...),
	)

	// 替换全局Logger
	zap.ReplaceGlobals(logger)

	// 记录初始化成功日志
	logger.Info("日志系统已初始化",
		zap.String("level", options.Level),
		zap.String("format", options.Format),
		zap.String("output", options.OutputPath),
		zap.String("error_output", options.ErrorOutputPath),
		zap.Int("max_size", options.MaxSize),
		zap.Int("max_backups", options.MaxBackups),
		zap.Int("max_age", options.MaxAge),
		zap.Bool("compress", options.Compress),
		zap.Bool("development", options.Development),
	)

	return nil
}

// Sync 同步日志缓冲区到输出
func Sync() {
	_ = zap.L().Sync()
}

// Logger 获取全局Logger
func Logger() *zap.Logger {
	return zap.L()
}

// WithLogger 在上下文中使用带有额外字段的Logger
func WithLogger(logger *zap.Logger, fields ...zap.Field) *zap.Logger {
	if len(fields) == 0 {
		return logger
	}
	return logger.With(fields...)
}

// NewFileRotateWriter 创建一个带日志轮转的writer
func NewFileRotateWriter(filename string, maxSize, maxBackups, maxAge int, compress bool) io.Writer {
	return &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    maxSize,
		MaxBackups: maxBackups,
		MaxAge:     maxAge,
		Compress:   compress,
	}
}
