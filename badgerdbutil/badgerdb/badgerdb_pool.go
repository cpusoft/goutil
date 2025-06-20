package badgerdb

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dgraph-io/badger/v4"
)

const (
	defaultMaxPoolSize = 50
	defaultMinPoolSize = 5
	maxIdleTime        = 5 * time.Minute
	cleanupInterval    = 30 * time.Second
)

var (
	ErrPoolClosed    = errors.New("pool is closed")
	ErrPoolExhausted = errors.New("connection pool exhausted")
	ErrInvalidConn   = errors.New("invalid connection")
)

// Pool 连接池结构
type Pool struct {
	mu          sync.RWMutex
	initialized uint32
	closed      uint32

	// 主数据库连接
	mainDB *badger.DB

	// 连接池配置
	maxSize     int32
	minSize     int32
	currentSize int32

	// 空闲连接管理
	idle chan *poolConn

	// 清理定时器
	cleanupTimer *time.Ticker

	// 数据库配置
	options badger.Options
}

// poolConn 池化连接
type poolConn struct {
	db       BadgeDB
	lastUsed time.Time
	inUse    bool
}

// NewPool 创建新的连接池
func NewPool(opts badger.Options) (*Pool, error) {
	mainDB, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	p := &Pool{
		mainDB:       mainDB,
		maxSize:      defaultMaxPoolSize,
		minSize:      defaultMinPoolSize,
		options:      opts,
		idle:         make(chan *poolConn, defaultMaxPoolSize),
		cleanupTimer: time.NewTicker(cleanupInterval),
	}

	// 预热连接池
	for i := int32(0); i < p.minSize; i++ {
		conn := &poolConn{
			db:       NewBadgeDBWithDB(mainDB),
			lastUsed: time.Now(),
		}
		p.idle <- conn
		atomic.AddInt32(&p.currentSize, 1)
	}

	atomic.StoreUint32(&p.initialized, 1)

	// 启动清理协程
	go p.cleanup()

	return p, nil
}

// cleanup 定期清理空闲连接
func (p *Pool) cleanup() {
	for range p.cleanupTimer.C {
		if atomic.LoadUint32(&p.closed) == 1 {
			return
		}

		p.mu.Lock()
		currentIdle := len(p.idle)
		if currentIdle <= int(p.minSize) {
			p.mu.Unlock()
			continue
		}

		var cleaned int
		for i := 0; i < currentIdle-int(p.minSize); i++ {
			select {
			case conn := <-p.idle:
				if time.Since(conn.lastUsed) > maxIdleTime {
					atomic.AddInt32(&p.currentSize, -1)
					cleaned++
					continue
				}
				p.idle <- conn
			default:
				break
			}
		}
		p.mu.Unlock()
	}
}

// Get 获取连接
func (p *Pool) Get() (*poolConn, error) {
	if atomic.LoadUint32(&p.closed) == 1 {
		return nil, ErrPoolClosed
	}

	// 先尝试从空闲连接获取
	select {
	case conn := <-p.idle:
		conn.inUse = true
		conn.lastUsed = time.Now()
		return conn, nil
	default:
		// 空闲连接池为空，尝试创建新连接
		if atomic.LoadInt32(&p.currentSize) < p.maxSize {
			p.mu.Lock()
			if atomic.LoadInt32(&p.currentSize) < p.maxSize {
				conn := &poolConn{
					db:       NewBadgeDBWithDB(p.mainDB),
					lastUsed: time.Now(),
					inUse:    true,
				}
				atomic.AddInt32(&p.currentSize, 1)
				p.mu.Unlock()
				return conn, nil
			}
			p.mu.Unlock()
		}
	}

	// 尝试等待空闲连接
	select {
	case conn := <-p.idle:
		conn.inUse = true
		conn.lastUsed = time.Now()
		return conn, nil
	case <-time.After(time.Second):
		return nil, ErrPoolExhausted
	}
}

// Put 归还连接
func (p *Pool) Put(conn *poolConn) error {
	if conn == nil {
		return ErrInvalidConn
	}

	if atomic.LoadUint32(&p.closed) == 1 {
		return ErrPoolClosed
	}

	conn.inUse = false
	conn.lastUsed = time.Now()

	select {
	case p.idle <- conn:
		return nil
	default:
		// 如果空闲队列已满，直接丢弃
		atomic.AddInt32(&p.currentSize, -1)
		return nil
	}
}

// Close 关闭连接池
func (p *Pool) Close() error {
	if !atomic.CompareAndSwapUint32(&p.closed, 0, 1) {
		return ErrPoolClosed
	}

	p.cleanupTimer.Stop()

	p.mu.Lock()
	defer p.mu.Unlock()

	// 清空空闲连接
	close(p.idle)
	for range p.idle {
		atomic.AddInt32(&p.currentSize, -1)
	}

	// 关闭主数据库连接
	if p.mainDB != nil {
		if err := p.mainDB.Close(); err != nil {
			return err
		}
		p.mainDB = nil
	}

	return nil
}

// Stats 获取连接池状态
func (p *Pool) Stats() PoolStats {
	return PoolStats{
		CurrentSize: atomic.LoadInt32(&p.currentSize),
		IdleSize:    int32(len(p.idle)),
		MaxSize:     p.maxSize,
		MinSize:     p.minSize,
	}
}

// PoolStats 连接池统计信息
type PoolStats struct {
	CurrentSize int32
	IdleSize    int32
	MaxSize     int32
	MinSize     int32
}

// WithConn 使用连接执行操作
func (p *Pool) WithConn(fn func(BadgeDB) error) error {
	conn, err := p.Get()
	if err != nil {
		return err
	}
	defer p.Put(conn)

	return fn(conn.db)
}

// WithConnTxn 在事务中执行操作
func (p *Pool) WithConnTxn(fn func(*badger.Txn) error) error {
	conn, err := p.Get()
	if err != nil {
		return err
	}
	defer p.Put(conn)

	return conn.db.RunWithTxn(fn)
}
