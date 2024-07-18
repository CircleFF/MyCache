package consistent_hash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

type HashFunc func(data []byte) uint32

// Hash 一致性哈希算法
type Hash struct {
	// hash 映射函数
	hash HashFunc
	// replicas 每个节点虚拟出的节点个数
	replicas int
	// virtualNodes 排序过的虚拟节点
	virtualNodes []int
	// nodeMap 虚拟与真实节点的映射
	nodeMap map[int]string
}

func NewHash(repl int, f HashFunc) *Hash {
	h := &Hash{
		hash:         f,
		replicas:     repl,
		virtualNodes: []int{},
		nodeMap:      map[int]string{},
	}

	if h.hash == nil {
		h.hash = crc32.ChecksumIEEE
	}
	return h
}

// AddNode 添加真实节点，创建虚拟节点
func (h *Hash) AddNode(nodes ...string) {

	for _, node := range nodes {
		for i := 0; i < h.replicas; i++ {
			vNode := int(h.hash([]byte(strconv.Itoa(i) + node)))
			h.virtualNodes = append(h.virtualNodes, vNode)
			h.nodeMap[vNode] = node
		}
	}

	sort.Ints(h.virtualNodes)
}

// GetNode 获取真实节点
func (h *Hash) GetNode(key string) string {
	if len(key) == 0 {
		return ""
	}

	vNode := int(h.hash([]byte(key)))
	idx := sort.Search(len(h.virtualNodes), func(i int) bool {
		return h.virtualNodes[i] >= vNode
	})
	// "n" means "0"
	realNode := h.nodeMap[h.virtualNodes[idx%len(h.virtualNodes)]]

	return realNode
}

// RemoveNode 节点下线
func (h *Hash) RemoveNode(nodes ...string) {
	delVNodes := make(map[int]struct{})
	newVNodes := []int{}
	for _, node := range nodes {
		for i := 0; i < h.replicas; i++ {
			vNode := int(h.hash([]byte(strconv.Itoa(i) + node)))
			delVNodes[vNode] = struct{}{}
			delete(h.nodeMap, vNode)
		}
	}
	for _, v := range h.virtualNodes {
		if _, ok := delVNodes[v]; !ok {
			newVNodes = append(newVNodes, v)
		}
	}
	h.virtualNodes = newVNodes
}
