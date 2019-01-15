package rewledis

import (
	"context"
	"errors"
	"io"
	"sync/atomic"
	"time"

	"github.com/gomodule/redigo/redis"
	"golang.org/x/sync/semaphore"
)

// Error variables related to SubPool.
var (
	ErrUnsupportedSubPoolConnection = errors.New("connection returned by SubPool is unsupported by this operation")
)

var _ redis.Conn = &subPoolConn{}

type subPoolConn struct {
	redis.Conn
	subPool *SubPool
}

func (s *subPoolConn) Close() error {
	if s.Conn == nil {
		return nil
	}

	err := s.Conn.Close()
	if s.subPool.MaxActive != 0 {
		s.subPool.semaphore.Release(1)
	}
	s.Conn = nil
	return err
}

const (
	subPoolStateUninitialised uint64 = iota
	subPoolStateInitialising
	subPoolStateInitialised
)

// SubPool represents a fixed-capacity part of an existing redis.Pool.
// The SubPool allows a maximum of MaxActive connections to be in use by its
// consumers at any point in time.
//
// The type mimics redis.Pool behaviour and should be initialised using struct
// initialisation.
//
//     subPool := SubPool{
//         Pool:      pool,
//         MaxActive: 5,
//     }
//
// Close() and further methods are purposefully omitted, the underlying
// redis.Pool is considered the owner and manager of the connection
// resources.
type SubPool struct {
	Pool      *redis.Pool
	MaxActive int

	// state is used to ensure that semaphore is initialised only once and
	// only by a single goroutine.
	state     uint64
	semaphore *semaphore.Weighted
}

func (s *SubPool) lazyInit() {
	state := atomic.LoadUint64(&s.state)
	if state == subPoolStateInitialised {
		return
	}

	if state == subPoolStateUninitialised &&
		atomic.CompareAndSwapUint64(&s.state, subPoolStateUninitialised, subPoolStateInitialising) {
		s.semaphore = semaphore.NewWeighted(int64(s.MaxActive))
		atomic.StoreUint64(&s.state, subPoolStateInitialised)
	} else {
		for {
			time.Sleep(time.Millisecond)
			state = atomic.LoadUint64(&s.state)
			if state == subPoolStateInitialised {
				break
			}
		}
	}
}

// Get retrieves a connection from the pool. The connection must be closed.
func (s *SubPool) Get() redis.Conn {
	s.lazyInit()

	ctx, cancel := context.WithCancel(context.Background())
	s.semaphore.Acquire(ctx, 1)
	cancel()

	poolConn := s.Pool.Get()
	return &subPoolConn{
		Conn:    poolConn,
		subPool: s,
	}
}

// GetContext retrieves a connection from the pool. GetContext respects the
// passed in context while retrieving the connection. The connection must be
// closed.
func (s *SubPool) GetContext(ctx context.Context) (redis.Conn, error) {
	var err error
	var poolConn redis.Conn

	s.lazyInit()

	if s.MaxActive != 0 {
		s.semaphore.Acquire(ctx, 1)
	}
	poolConn, err = s.Pool.GetContext(ctx)
	if err != nil {
		return nil, err
	}

	return &subPoolConn{
		Conn:    poolConn,
		subPool: s,
	}, nil
}

type closerConn struct {
	redis.Conn
	closer io.Closer
}

func (c *closerConn) Close() error {
	if c.Conn == nil {
		return nil
	}
	err := c.closer.Close()
	c.Conn = nil
	return err
}

// getRaw returns a raw, unwrapped connection from the sub pool.
// The unwrapping is performed through the UNSAFE SELF rewledis command.
func (s *SubPool) getRaw(ctx context.Context) (redis.Conn, error) {
	poolConn, err := s.GetContext(ctx)
	if err != nil {
		return nil, err
	}

	connIntf, err := poolConn.Do("UNSAFE", "SELF")
	if err != nil {
		poolConn.Close()
		return nil, err
	}

	conn, ok := connIntf.(redis.Conn)
	if !ok {
		poolConn.Close()
		return nil, ErrUnsupportedSubPoolConnection
	}

	return &closerConn{
		Conn:   conn,
		closer: poolConn,
	}, nil
}
