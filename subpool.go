package rewledis

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/gomodule/redigo/redis"
	"golang.org/x/sync/semaphore"
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
	s.subPool.semaphore.Release(1)
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
// redis.Pool is considered the owner and managers of the connection
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
	ctx := context.Background()
	s.semaphore.Acquire(ctx, 1)
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
	s.lazyInit()
	s.semaphore.Acquire(ctx, 1)
	poolConn, err := s.Pool.GetContext(ctx)
	if err != nil {
		return nil, err
	}
	return &subPoolConn{
		Conn:    poolConn,
		subPool: s,
	}, nil
}
