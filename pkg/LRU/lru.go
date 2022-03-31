// Package LRU 缓存淘汰策略
//最近最少使用，相对于仅考虑时间因素的 FIFO 和仅考虑访问频率的 LFU ，LRU 算法可以认为是相对平衡的一种淘汰算法。
//LRU 认为，如果数据最近被访问过，那么将来被访问的概率也会更高。LRU 算法的实现非常简单，维护一个队列，如果某条记
//录被访问了，则移动到队尾，那么队首则是最近最少访问的数据，淘汰该条记录即可。
package LRU

import (
	"container/list"
)

type Cache struct {
	maxBytes  int64
	nbytes    int64
	ll        *list.List
	cache     map[string]*list.Element
	OnEvicted func(key string, value Value)
}

func NewCache(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

type entry struct {
	key   string
	value Value
}

//Value 使用Len()计算他需要的字节
type Value interface {
	Len() int
}

//Get 查找一个key对应的value
func (c *Cache) Get(key string) (value Value, ok bool) {
	//从字典中找到对应的双向链表
	//将该节点移动到队尾
	if ele, ok := c.cache[key]; ok { //如果对应的链表节点存在 将其移动到队尾并返回查找的值
		c.ll.MoveToFront(ele) //将链表结点ele移动到队尾（双向链表队首队尾是相对的）
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return
}

//RemoveOldest 缓存淘汰 移除最近最少访问的节点
func (c *Cache) RemoveOldest() {
	ele := c.ll.Back() //取到队首节点 将其删除
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)                                //从map中删除对应节点的映射关系
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len()) //更新当前所用内存
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value) //回调函数
		}
	}
}

//Add 新增缓存
func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		ele1 := c.ll.PushFront(&entry{key, value})
		c.cache[key] = ele1
		c.nbytes += int64(len(key)) + int64(value.Len())
	}
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest()
	}
}

//Len 缓存条目的数量
func (c *Cache) Len() int {
	return c.ll.Len()
}
