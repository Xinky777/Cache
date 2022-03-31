package cacheGo

//并发控制

import (
	"Cache/pkg/LRU"
	"sync"
)

type cache struct {
	mu         sync.Mutex
	lru        *LRU.Cache
	cacheBytes int64
}

func (c *cache) add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	//判断c.lru是否为nil 如果为nil再开辟空间创造实例 该方法叫延迟初始化
	//延迟初始化意味着该对象的创建会延迟到第一次使用该对象时
	//主要用于提升性能 减少程序内存需求
	if c.lru == nil {
		c.lru = LRU.NewCache(c.cacheBytes, nil)
	}
	c.lru.Add(key, value)
}

func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		return
	}
	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), ok
	}
	return
}
