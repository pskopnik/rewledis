package rewledis

import (
	"github.com/gomodule/redigo/redis"
)

func NewPool(poolConfig *PoolConfig) *redis.Pool {
	rewriter := &Rewriter{}
	return rewriter.NewPrimaryPool(poolConfig)
}
