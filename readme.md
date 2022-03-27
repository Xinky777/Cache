支持的特性：
1.单机缓存和基于HTTP的分布式缓存
2.最近最少访问（Least Recently Used,LRU）缓存策略
3.使用Go锁机制防止缓存击穿
4.使用一致性哈希选择节点，实现负载均衡
5.使用protobuf优化节点之间的二进制通信