package rewledis

import (
	"time"

	"github.com/gomodule/redigo/redis"
)

//go:generate confions config PoolConfig

// PoolConfig contains all configuration options for a redigo pool instance.
// The fields contained mirror the redis.Pool type of redigo.
type PoolConfig struct {
	// Dial is an application supplied function for creating and configuring a
	// connection. Dial must return a connection to a LedisDB server.
	Dial func() (redis.Conn, error)

	// TestOnBorrow is an optional application supplied function for checking the
	// health of an idle connection before the connection is used again by the
	// application. Argument c is a wrapped (rewriting) connection emulating
	// Redis semantics. Argument t is the time that the connection was returned
	// to the pool. If the function returns an error, then the connection is
	// closed.
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

// NewPool is a convenience function creating a new Pool returning rewriting
// connections.
// poolConfig and internalMaxActive are passed on to
// (*Rewriter).NewPrimaryPool().
func NewPool(poolConfig *PoolConfig, internalMaxActive int) *redis.Pool {
	rewriter := &Rewriter{}
	return rewriter.NewPrimaryPool(poolConfig, internalMaxActive)
}
