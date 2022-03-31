package cacheGo

//PeerPicker 根据传入的key选择相应的节点PeerGetter
type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

//PeerGetter 从对应的group中查找缓存值 对应流程图中的http客户端
type PeerGetter interface {
	Get(group string, key string) ([]byte, error)
}
