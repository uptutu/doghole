package config

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"entgo.io/ent/dialect"
	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	// _GlobalConfig 全局配置实例
	_GlobalConfig *Config
)

// GetGlobalConfig 获取全局配置
func GetGlobalConfig() *Config {
	if _GlobalConfig == nil {
		_GlobalConfig = NewConfig()
	}
	return _GlobalConfig
}

// SetGlobalConfig 设置全局配置
func SetGlobalConfig(c *Config) {
	if c == nil {
		panic("无法设置全局配置为nil")
	}
	_GlobalConfig = c
}

// Config 应用配置结构体
type Config struct {
	Server ServerConfig `json:"server" mapstructure:"server"` // 服务器配置
	DB     DBConfig     `json:"db" mapstructure:"db"`         // 数据库配置
	Logger LoggerConfig `json:"logger" mapstructure:"logger"` // 日志配置
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port              int           `json:"port" mapstructure:"port"`                             // 服务器端口
	ReadTimeout       time.Duration `json:"read_timeout" mapstructure:"read_timeout"`             // 读取超时
	WriteTimeout      time.Duration `json:"write_timeout" mapstructure:"write_timeout"`           // 写入超时
	IdleTimeout       time.Duration `json:"idle_timeout" mapstructure:"idle_timeout"`             // 空闲超时
	ShutdownTimeout   time.Duration `json:"shutdown_timeout" mapstructure:"shutdown_timeout"`     // 关闭超时
	EnableCompression bool          `json:"enable_compression" mapstructure:"enable_compression"` // 启用压缩
	EnablePrefork     bool          `json:"enable_prefork" mapstructure:"enable_prefork"`         // 启用预分叉
}

// DBConfig 数据库配置
type DBConfig struct {
	WriteDB *DB `json:"write_db" mapstructure:"write_db"` // 写入数据库配置
	ReadDB  *DB `json:"read_db" mapstructure:"read_db"`   // 读取数据库配置
	DB      *DB `json:"db" mapstructure:"db"`             // 单一数据库配置
}

// LoggerConfig 日志配置
type LoggerConfig struct {
	Level     string `json:"level" mapstructure:"level"`           // 日志级别
	Format    string `json:"format" mapstructure:"format"`         // 日志格式
	Outfile   string `json:"outfile" mapstructure:"outfile"`       // 输出文件路径
	ChuckSize int    `json:"chuck_size" mapstructure:"chuck_size"` // 日志切割大小
}

// DB 数据库连接配置
type DB struct {
	Driver   string `json:"driver" mapstructure:"driver"`     // 数据库驱动
	Host     string `json:"host" mapstructure:"host"`         // 主机地址
	Port     int    `json:"port" mapstructure:"port"`         // 端口
	Username string `json:"username" mapstructure:"username"` // 用户名
	Password string `json:"password" mapstructure:"password"` // 密码
	Database string `json:"database" mapstructure:"database"` // 数据库名
	SSLMode  string `json:"ssl_mode" mapstructure:"ssl_mode"` // SSL模式
}

// ToDialect 转换为ent方言
func (db *DB) ToDialect() string {
	if db == nil {
		return ""
	}
	return db.Driver
}

// ToDNS 生成数据库连接字符串
func (db *DB) ToDNS() string {
	if db == nil {
		return ""
	}

	switch db.Driver {
	case dialect.Postgres:
		return fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=%s",
			db.Host, db.Port, db.Username, db.Database, db.Password, db.SSLMode)
	case dialect.MySQL:
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=True",
			db.Username, db.Password, db.Host, db.Port, db.Database)
	case dialect.SQLite:
		return db.Database
	default:
		return ""
	}
}

// NewConfig 创建一个带有默认值的新配置
func NewConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port:              8080,
			ReadTimeout:       30 * time.Second,
			WriteTimeout:      30 * time.Second,
			IdleTimeout:       120 * time.Second,
			ShutdownTimeout:   5 * time.Second,
			EnableCompression: true,
			EnablePrefork:     false,
		},
		Logger: LoggerConfig{
			Level:     "info",
			Format:    "json",
			Outfile:   "",
			ChuckSize: 100,
		},
	}
}

// ToPort 生成带冒号的端口字符串
func (c *Config) ToPort() string {
	return fmt.Sprintf(":%d", c.Server.Port)
}

// LoadSingleConfigFile 从单个配置文件加载配置
func (c *Config) LoadSingleConfigFile(filename string) error {
	filetype, err := assertFileType(filename)
	if err != nil {
		return errors.Wrap(err, "无法识别配置文件类型")
	}

	v := viper.New()
	v.SetConfigFile(filename)
	v.SetConfigType(filetype)

	if err := v.ReadInConfig(); err != nil {
		return errors.Wrap(err, "读取配置文件失败")
	}

	if err := v.Unmarshal(c); err != nil {
		return errors.Wrap(err, "解析配置文件失败")
	}

	// 监听配置文件变化
	v.WatchConfig()
	v.OnConfigChange(func(e fsnotify.Event) {
		zap.L().Info("配置文件已更改", zap.String("file", e.Name))
		if err := v.Unmarshal(c); err != nil {
			zap.L().Error("重新加载配置失败", zap.Error(err))
		}
	})

	return nil
}

// LoadConfigFromDirs 从多个目录加载配置
func (c *Config) LoadConfigFromDirs(filename string, dirs ...string) error {
	v := viper.New()

	for _, dir := range dirs {
		v.AddConfigPath(dir)
	}

	base := filepath.Base(filename)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)

	v.SetConfigName(name)
	v.SetConfigType(strings.TrimPrefix(ext, "."))

	if err := v.ReadInConfig(); err != nil {
		return errors.Wrap(err, "读取配置文件失败")
	}

	if err := v.Unmarshal(c); err != nil {
		return errors.Wrap(err, "解析配置文件失败")
	}

	// 监听配置文件变化
	v.WatchConfig()
	v.OnConfigChange(func(e fsnotify.Event) {
		zap.L().Info("配置文件已更改", zap.String("file", e.Name))
		if err := v.Unmarshal(c); err != nil {
			zap.L().Error("重新加载配置失败", zap.Error(err))
		}
	})

	return nil
}

// LoadEnvConfig 从环境变量加载配置
func (c *Config) LoadEnvConfig(prefix string) error {
	v := viper.New()
	v.SetEnvPrefix(prefix)
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 将环境变量绑定到配置结构
	bindEnvs(v, c)

	return nil
}

// bindEnvs 递归绑定环境变量
func bindEnvs(v *viper.Viper, iface interface{}, parts ...string) {
	// 此函数实现递归绑定环境变量
	// 在实际实现中，需要使用反射来遍历结构体字段
}

// assertFileType 检测文件类型
func assertFileType(filename string, allowedTypes ...string) (string, error) {
	ext := strings.TrimPrefix(filepath.Ext(filename), ".")

	if len(allowedTypes) == 0 {
		allowedTypes = []string{"yaml", "yml", "json", "toml", "ini"}
	}

	for _, t := range allowedTypes {
		if ext == t {
			return ext, nil
		}
	}

	return "", fmt.Errorf("不支持的文件类型: %s, 支持的类型: %v", ext, allowedTypes)
}
