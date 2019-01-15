package rewledis

import (
	"github.com/gomodule/redigo/redis"
)

type Rewriter struct {
	cache           Cache
	primaryPool     *redis.Pool
	internalSubPool SubPool
}

// NewPrimaryPool creates a new pool from config and uses the created pool as
// its primary pool. The primary pool is used for internal operations, at the
// moment only by a Resolver. Use internalMaxActive to set the maximum number
// of connections used for internal purposes. 0 means no limit.
//
// The primary pool is not changed if the primary pool of the Rewriter has
// already been set and this method is called again.
//
// The returned Pool yields wrapped connections emulating Redis semantics.
// Commands are rewritten using this Rewriter.
func (r *Rewriter) NewPrimaryPool(config *PoolConfig, internalMaxActive int) *redis.Pool {
	pool := &redis.Pool{
		Dial: func() (redis.Conn, error) {
			conn, err := config.Dial()
			if err != nil {
				return nil, err
			}

			return &LedisConn{
				rewriter: r,
				conn:     conn,
			}, nil
		},
		TestOnBorrow:    config.TestOnBorrow,
		MaxIdle:         config.MaxIdle,
		MaxActive:       config.MaxActive,
		IdleTimeout:     config.IdleTimeout,
		Wait:            config.Wait,
		MaxConnLifetime: config.MaxConnLifetime,
	}

	if r.primaryPool == nil {
		r.primaryPool = pool
		r.internalSubPool.Pool = pool
		r.internalSubPool.MaxActive = internalMaxActive
	}

	return pool
}

// NewPool creates and returns a new Pool of connections to a LedisDB server.
// The returned Pool yields wrapped connections emulating Redis semantics.
// Commands are rewritten using this Rewriter.
//
// In order for rewriting to work properly, the primary pool of the Rewriter
// must have been set. That means NewPrimaryPool has to have been called.
func (r *Rewriter) NewPool(config *PoolConfig) *redis.Pool {
	pool := &redis.Pool{
		Dial: func() (redis.Conn, error) {
			conn, err := config.Dial()
			if err != nil {
				return nil, err
			}

			return &LedisConn{
				rewriter: r,
				conn:     conn,
			}, nil
		},
		TestOnBorrow:    config.TestOnBorrow,
		MaxIdle:         config.MaxIdle,
		MaxActive:       config.MaxActive,
		IdleTimeout:     config.IdleTimeout,
		Wait:            config.Wait,
		MaxConnLifetime: config.MaxConnLifetime,
	}

	return pool
}

// Resolver constructs and returns a Resolver instance using this rewriter.
func (r *Rewriter) Resolver() Resolver {
	return Resolver{
		Cache:   &r.cache,
		SubPool: &r.internalSubPool,
	}
}

// WrapConn wraps a connection to a LedisDB server and returns a connection
// emulating Redis semantics.
// All commands issued on the returned connection are rewritten using this rewriter.
func (r *Rewriter) WrapConn(conn redis.Conn) *LedisConn {
	if ledisConn, ok := conn.(*LedisConn); ok {
		return &LedisConn{
			rewriter: r,
			conn:     ledisConn.conn,
		}
	}

	return &LedisConn{
		rewriter: r,
		conn:     conn,
	}
}

// Rewrite applies transformations for a single supplied command invocation.
func (r *Rewriter) Rewrite(commandName string, args ...interface{}) (SendLedisFunc, error) {
	command, err := RedisCommandFromName(commandName)
	if err != nil {
		return nil, err
	}

	return command.TransformFunc(r, command, args)
}
