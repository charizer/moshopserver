package cache

import (
	"github.com/patrickmn/go-cache"
	"time"
)

var MemCachePool *MemCache

func InitMemCache() {
	MemCachePool = NewMemCache()
}

type MemCache struct {
	c *cache.Cache
}

func NewMemCache() *MemCache {
	return &MemCache{
		c: cache.New(30 * time.Minute, 30 * time.Minute),
	}
}

func (m *MemCache) Set(k string, v interface{}, d int) {
	m.c.Set(k, v, time.Duration(d) * time.Minute)
}

func (m *MemCache) Get(k string) (interface{}, bool) {
	return m.c.Get(k)
}
