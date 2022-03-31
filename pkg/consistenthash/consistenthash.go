package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

//Hash 将字节映射到uint32
type Hash func(data []byte) uint32

//Map 包含所有的哈希key
type Map struct {
	hash     Hash           //哈希函数
	replicas int            //虚拟节点倍数
	keys     []int          //哈希环
	hashMap  map[int]string //虚拟节点与真实节点映射表 键是虚拟节点的哈希值 值是真实节点的名称
}

//New 创造一个Map实例 允许自定义虚拟节点倍数和hash函数
func New(replicas int, fn Hash) *Map {
	m := &Map{
		hash:     fn,
		replicas: replicas,
		hashMap:  make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

//Add 添加一些节点到哈希表中
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ { //每一个真实节点key 都对应创建m.replicas个虚拟节点
			//使用m.hash计算虚拟节点的哈希值
			hash := int(m.hash([]byte(strconv.Itoa(i) + key))) //虚拟节点的名称是 strconv.Itoa(i) + key
			//即通过添加编号的方式区分不同的虚拟节点
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key //添加虚拟节点与真实节点的映射关系
		}
	}
	//环上哈希值排序
	sort.Ints(m.keys)
}

func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}
	//计算key的哈希值
	hash := int(m.hash([]byte(key)))
	//二进制搜索适当的虚拟节点倍数
	//顺时针找到第一个匹配的虚拟节点的下标idx
	idx := sort.Search(len(m.keys), func(i int) bool {
		//如果 idx == len(m.keys) 说明应该选择m.keys[0],应为m.keys是一个环状结构 所以用取余数的方法来处理这种情况
		return m.keys[i] >= hash
	})
	//通过hashMap映射得到真实的节点
	return m.hashMap[m.keys[idx%len(m.keys)]]
}
