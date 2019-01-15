package rewledis

import (
	"github.com/gomodule/redigo/redis"
)

type Rewriter struct {
	cache         Cache
	ResolvingPool *redis.Pool
}

// NewPool creates and returns a new Pool of connections to a LedisDB server.
// The returned Pool yields wrapped connections emulating Redis semantics.
// Commands are rewritten using this Rewriter.
//
// In order for rewriting to work properly, the ResolvingPool of the Rewriter
// has to be instantiated.
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

// ResolvingPoolFromConfig instantiates a new ResolvingPool from the passed in
// PoolConfig.
func (r *Rewriter) ResolvingPoolFromConfig(config *PoolConfig) {
	r.ResolvingPool = &redis.Pool{
		Dial:            config.Dial,
		TestOnBorrow:    config.TestOnBorrow,
		MaxIdle:         config.MaxIdle,
		MaxActive:       config.MaxActive,
		IdleTimeout:     config.IdleTimeout,
		Wait:            config.Wait,
		MaxConnLifetime: config.MaxConnLifetime,
	}
}

// Resolver constructs and returns a Resolver instance using this rewriter.
func (r *Rewriter) Resolver() Resolver {
	return Resolver{
		Cache: &r.cache,
		Pool:  r.ResolvingPool,
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
