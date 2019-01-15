package rewledis

import (
	"time"
)

func (p *PoolConfig) CopyFrom(other *PoolConfig) {
	p.Dial = other.Dial
	p.TestOnBorrow = other.TestOnBorrow
	p.MaxIdle = other.MaxIdle
	p.MaxActive = other.MaxActive
	p.IdleTimeout = other.IdleTimeout
	p.Wait = other.Wait
	p.MaxConnLifetime = other.MaxConnLifetime
}

func (p *PoolConfig) Merge(other *PoolConfig) *PoolConfig {
	if other.Dial != nil {
		p.Dial = other.Dial
	}
	if other.TestOnBorrow != nil {
		p.TestOnBorrow = other.TestOnBorrow
	}
	if other.MaxIdle != 0 {
		p.MaxIdle = other.MaxIdle
	}
	if other.MaxActive != 0 {
		p.MaxActive = other.MaxActive
	}
	if other.IdleTimeout != time.Duration(0) {
		p.IdleTimeout = other.IdleTimeout
	}
	if other.Wait != false {
		p.Wait = other.Wait
	}
	if other.MaxConnLifetime != time.Duration(0) {
		p.MaxConnLifetime = other.MaxConnLifetime
	}

	return p
}

func (p *PoolConfig) Clone() *PoolConfig {
	config := &PoolConfig{}
	config.CopyFrom(p)
	return config
}
