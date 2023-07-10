package consistenthash

import (
	"hash/crc32"
	"sort"
)

//存储所有节点的信息（集群）和所有节点的一致性哈希的信息
type HashFunc func(data []byte) uint32
type NodeMap struct {
	hashFunc    HashFunc       //hash函数
	nodeHashs   []int          //保存节点的哈希值（例如12354，2144，58885，42512节点信息，并且是按照顺序保存）
	nodehashMap map[int]string //key就是上面哈希值，value就是节点具体位置。这个主要记录该节点放在具体的那个位置上
	//A:12354，B:2144，C:58885，D:42512
}

//初始化HashFunc结构体，而且用户可以自己定义hashFunc，如果用户没有传入则默认使用crc32.ChecksumIEEE
func NewNodeMap(fn HashFunc) *NodeMap {
	m := &NodeMap{
		hashFunc:    fn,
		nodehashMap: make(map[int]string),
	}
	if m.hashFunc == nil {
		m.hashFunc = crc32.ChecksumIEEE
	}
	return m
}

//写一个方法：判断整个一致性哈希或整个集群它的节点是不是为空
func (m *NodeMap) IsEmpty() bool {
	return len(m.nodeHashs) == 0
}

//将一个新的节点加入到一致性哈希的哈希环中
//传一组k,k可能是节点名字或者IP，只要唯一确定该节点就可以
/*
	1. 首先进行哈希
	2. 排序（方便后面新来的k,v落在那两个节点中间）
	3. 为了方便找到该哈希值落在具体哪个节点（通过Map将hash节点和节点名对应）
*/
func (m *NodeMap) AddNode(keys ...string) {
	for _, key := range keys {
		if key == "" {
			continue
		}
		hash := int(m.hashFunc([]byte(key))) //进行hashFunc运算计算hash值
		m.nodeHashs = append(m.nodeHashs, hash)
		m.nodehashMap[hash] = key //格式：12345->节点1
	}
	sort.Ints(m.nodeHashs)
}

//来一个k,v具体去哪一个节点。函数传入key返回节点具体地址
func (m *NodeMap) PickNode(key string) string {
	//首先判断该集群有节点吗
	if m.IsEmpty() {
		return ""
	}
	hash := int(m.hashFunc([]byte(key)))
	//去nodeHashs里面搜索，看看是在哪俩个节点之间
	idx := sort.Search(len(m.nodeHashs), func(i int) bool {
		return m.nodeHashs[i] >= hash
	})
	//判断数组长度加入key的hash值是6000，超过了最后最后一个节点，那他就放在第一个节点上
	if idx == len(m.nodeHashs) {
		idx = 0
	}
	return m.nodehashMap[m.nodeHashs[idx]] //用idx找到它节点对应的Hash，在使用该Hash找到对应的节点地址
}
