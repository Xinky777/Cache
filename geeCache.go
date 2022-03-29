// Package Cache 负责与外部交互，控制缓存存储和获取的主流程
package Cache

import (
	"fmt"
	"log"
	"sync"
)

//Group 是一个缓存的命名空间
type Group struct {
	name      string //每个group拥有唯一的名称 name
	getter    Getter //缓存未命中时获取原数据的回调
	mainCache cache  //并发缓存
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

//NewGroup 构造函数 实例化Group 并将group存储在全局变量groups
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		//当函数F调用panic时，F的正常执行就会立刻停止。F中defer的所有函数先入后出执行后，F返回给其调用者G。
		//G如同F一样行动，层层返回，直到该Go程中所有函数都按相反的顺序停止执行。之后，程序被终止，而错误情况
		//会被报告，包括引发该恐慌的实参值，此终止序列称为恐慌过程。
		panic("nil Getter") //恐慌过程
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
	}
	groups[name] = g
	return g
}

//GetGroup 返回之前使用的NewGroup创建的命名组
//如果没有这样的组 则返回nil
func GetGroup(name string) *Group {
	mu.RLock() //只用了只读锁 因为不涉及任何冲突变量的写操作
	g := groups[name]
	mu.RUnlock()
	return g
}

//Get 实现readme里面流程的（1）和（3）
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required") //Errorf 根据格式说明符进行格式化，并将字符串作为满足错误的值返回
	}
	//流程（1）：从mainCache中查找缓存 如果存在则存在返回值
	if v, ok := g.mainCache.get(key); ok {
		log.Println("[GeeCache] hit")
		return v, nil
	}
	//流程（3）：如果缓存不存在 则调用load方法
	return g.load(key)
}

//load load调用getLocally（分布式场景下会调用getFromPeer 从其他节点获取）
func (g *Group) load(key string) (value ByteView, err error) {
	return g.getLocally(key)
}

//getLocally 调用用户回调函数g.getter.Get()获取源数据 并且将其利用populateCache方法添加到缓存mainCache中
func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

//populateCache 将源数据添加到缓存mainCache中
func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}

//Getter 为键加载数据
type Getter interface {
	Get(key string) ([]byte, error)
}

//GetterFunc 使用函数实现Getter
type GetterFunc func(key string) ([]byte, error)

//Get 实现Getter接口函数
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}
