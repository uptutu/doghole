package conn

import (
	"context"
	"sync"
	"time"

	"doghole/ent"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

var (
	_writeConn   *ent.Client
	_readConn    *ent.Client
	_connMutex   sync.RWMutex
	_initialized bool
)

// DBManager 数据库连接管理器
type DBManager struct {
	writeClient *ent.Client
	readClient  *ent.Client
	logger      *zap.Logger
}

// NewDBManager 创建数据库连接管理器
func NewDBManager(options ...func(*DBManager)) (*DBManager, error) {
	manager := &DBManager{
		logger: zap.L(),
	}

	for _, option := range options {
		option(manager)
	}

	// 验证连接
	if manager.writeClient == nil && manager.readClient == nil {
		return nil, errors.New("必须至少提供一个读取或写入数据库连接")
	}

	// 如果只提供了一个连接，则读写共用同一个连接
	if manager.writeClient == nil {
		manager.writeClient = manager.readClient
	}
	if manager.readClient == nil {
		manager.readClient = manager.writeClient
	}

	return manager, nil
}

// WithWriteClient 设置写入客户端
func WithWriteClient(client *ent.Client) func(*DBManager) {
	return func(m *DBManager) {
		m.writeClient = client
	}
}

// WithReadClient 设置读取客户端
func WithReadClient(client *ent.Client) func(*DBManager) {
	return func(m *DBManager) {
		m.readClient = client
	}
}

// WithLogger 设置日志记录器
func WithLogger(logger *zap.Logger) func(*DBManager) {
	return func(m *DBManager) {
		m.logger = logger
	}
}

// Initialize 初始化数据库连接
func Initialize(ctx context.Context, writeClient, readClient *ent.Client) error {
	_connMutex.Lock()
	defer _connMutex.Unlock()

	if _initialized {
		return errors.New("数据库连接已初始化")
	}

	if writeClient == nil && readClient == nil {
		return errors.New("必须至少提供一个读取或写入数据库连接")
	}

	// 如果只提供了一个连接，则读写共用同一个连接
	if writeClient == nil {
		writeClient = readClient
	}
	if readClient == nil {
		readClient = writeClient
	}

	// 测试连接
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := writeClient.Schema.Create(timeoutCtx); err != nil {
		return errors.Wrap(err, "初始化写入数据库架构失败")
	}

	_writeConn = writeClient
	_readConn = readClient
	_initialized = true

	return nil
}

// SetWriteConn 设置写入连接
func SetWriteConn(conn *ent.Client) {
	_connMutex.Lock()
	defer _connMutex.Unlock()
	_writeConn = conn
}

// SetReadConn 设置读取连接
func SetReadConn(conn *ent.Client) {
	_connMutex.Lock()
	defer _connMutex.Unlock()
	_readConn = conn
}

// Writer 获取写入连接
func Writer() *ent.Client {
	_connMutex.RLock()
	defer _connMutex.RUnlock()
	return _writeConn
}

// Reader 获取读取连接
func Reader() *ent.Client {
	_connMutex.RLock()
	defer _connMutex.RUnlock()
	return _readConn
}

// Close 关闭所有数据库连接
func Close() {
	_connMutex.Lock()
	defer _connMutex.Unlock()

	if _writeConn != nil {
		if err := _writeConn.Close(); err != nil {
			zap.L().Error("关闭写入数据库连接失败", zap.Error(err))
		}
	}

	// 如果读写连接不是同一个对象，则关闭读连接
	if _readConn != nil && _readConn != _writeConn {
		if err := _readConn.Close(); err != nil {
			zap.L().Error("关闭读取数据库连接失败", zap.Error(err))
		}
	}

	_writeConn = nil
	_readConn = nil
	_initialized = false
}

// WithTx 在事务中执行函数
func WithTx(ctx context.Context, fn func(tx *ent.Tx) error) error {
	tx, err := Writer().Tx(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if v := recover(); v != nil {
			_ = tx.Rollback()
			panic(v)
		}
	}()

	if err := fn(tx); err != nil {
		if rerr := tx.Rollback(); rerr != nil {
			zap.L().Error("回滚事务失败", zap.Error(rerr))
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "提交事务失败")
	}

	return nil
}
