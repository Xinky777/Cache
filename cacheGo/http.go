package cacheGo

import (
	"Cache/pkg/consistenthash"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const (
	defaultBasePath = "/_geecache/"
	defaultReplicas = 50
)

//HTTPPool 为HTTP对等点池实现PeerPicker
type HTTPPool struct {
	self        string                 //记录自己的地址 包括主机名（IP）和端口
	basePath    string                 //节点通讯地址的前缀 默认是_geecache
	mu          sync.Mutex             //读写锁 保护peers和httpGetters
	peers       *consistenthash.Map    //成员变量，类型是一致性哈希 用来根据具体的key选择节点
	httpGetters map[string]*httpGetter //映射远程节点和对应的httpGetter
}

// NewHTTPPool 实现一个HTTP对等点池
func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

// Log 带有服务器名称的信息
func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[server %s]%s", p.self, fmt.Sprintf(format, v...))
}

//ServeHTTP 处理所有http请求
func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}
	p.Log("%s %s", r.Method, r.URL.Path)
	// /<basepath>/<groupname>/<key>required
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	groupName := parts[0]
	key := parts[1]
	group := GetGroup(groupName) //得到group实例
	if group == nil {
		http.Error(w, "no such group"+groupName, http.StatusNotFound)
		return
	}
	view, err := group.Get(key) //获取缓存数据
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(view.ByteSlice()) //将缓存值作为httpResponse的body返回
}

// Set updates the pool's list of peers.
func (p *HTTPPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers = consistenthash.New(defaultReplicas, nil)
	p.peers.Add(peers...)
	p.httpGetters = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		p.httpGetters[peer] = &httpGetter{baseURL: peer + p.basePath}
	}
}

// PickPeer picks a peer according to key
func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		p.Log("Pick peer %s", peer)
		return p.httpGetters[peer], true
	}
	return nil, false
}

var _ PeerPicker = (*HTTPPool)(nil)

//httpGetter 表示要访问的远程节点的地址
type httpGetter struct {
	baseURL string
}

func (h *httpGetter) Get(group string, key string) ([]byte, error) {
	u := fmt.Sprintf(
		"%v%v%v",
		h.baseURL,
		url.QueryEscape(group),
		url.QueryEscape(key),
	)
	res, err := http.Get(u) //使用http.Get方法获取返回值 并转换为[]bytes 类型
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned:%v", res.Status)
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body:%v", err)
	}
	return bytes, err
}
