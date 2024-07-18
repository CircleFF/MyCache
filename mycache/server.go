package mycache

// NodePicker 通过一致性哈希算法根据 key 选择节点
type NodePicker interface {
	PickNode(key string) (DataGetter, bool)
}

// DataGetter 从节点获取值
type DataGetter interface {
	GetData(group, key string) ([]byte, error)
}
