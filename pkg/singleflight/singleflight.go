package singleflight

import "sync"

//call 表示正在进行中 或者已经结束的请求
//使用锁避免重复写入
type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

//Group singleflight主数据结构 关机不同key的请求
type Group struct {
	mu sync.Mutex //保护m
	m  map[string]*call
}

//Do 针对相同的key 无论被调用多少次 函数fn只会被调用一次
//等待fn调用结束 返回返回值或错误
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	//延迟初始化
	//一个对象的延迟初始化意味着该对象的创建将会延迟至第一次使用该对象时。主要用于提高性能，并减少程序内存要求
	if g.m == nil {
		g.m = make(map[string]*call)
	}

	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()         //如果请求正在进行中 等待
		return c.val, c.err //请求结束 返回结果
	}

	c := new(call)
	c.wg.Add(1)  //发起请求前加锁
	g.m[key] = c //添加到g.m 表面key已经有对应的请求在处理
	g.mu.Unlock()

	c.val, c.err = fn() //调用函数 fn 发起请求
	c.wg.Done()         //请求结束

	g.mu.Lock()
	delete(g.m, key) //更新g.m
	g.mu.Unlock()

	return c.val, c.err //返回结果
}
