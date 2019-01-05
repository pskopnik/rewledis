package rewledis

import (
	"time"

	"github.com/gomodule/redigo/redis"
)

type PoolConfig struct {
	// Dial is an application supplied function for creating and configuring a
	// connection. Dial must return a connection to a LedisDB server.
	Dial func() (redis.Conn, error)

	// TestOnBorrow is an optional application supplied function for checking
	// the health of an idle connection before the connection is used again by
	// the application. Argument c is an unwrapped (non-rewriting) connection
	// to a LedisDB server as returned by Dial. Argument t is the time that
	// the connection was returned to the pool. If the function returns an
	// error, then the connection is closed.
	TestOnBorrow func(c redis.Conn, t time.Time) error

	// Maximum number of idle connections in the pool.
	MaxIdle int

	// MaxActive is the maximum number of connections allocated by the pool at
	// a given time. When zero, there is no limit on the number of connections
	// in the pool.
	MaxActive int

	// IdleTimeout is the duration after which to close idle connections. If
	// the value is zero, then idle connections are not closed. Applications
	// should set the timeout to a value less than the server's timeout.
	IdleTimeout time.Duration

	// If Wait is true and the pool is at the MaxActive limit, then Get() waits
	// for a connection to be returned to the pool before returning.
	Wait bool

	// Close connections older than this duration. If the value is zero, then
	// the pool does not close connections based on age.
	MaxConnLifetime time.Duration
}

type Rewriter struct {
	cache            Cache
	resolvingSubPool SubPool
	primaryPool      *redis.Pool
}

func (r *Rewriter) NewPrimaryPool(config *PoolConfig) *redis.Pool {
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
		r.resolvingSubPool.Pool = pool
		r.resolvingSubPool.MaxActive = 10
	}

	return pool
}

func (r *Rewriter) Resolver() Resolver {
	return Resolver{
		Cache:   &r.cache,
		SubPool: &r.resolvingSubPool,
	}
}

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

func (r *Rewriter) Rewrite(commandName string, args ...interface{}) (SendLedisFunc, error) {
	command, err := RedisCommandFromName(commandName)
	if err != nil {
		return nil, err
	}

	return command.TransformFunc(r, command, args)
}
