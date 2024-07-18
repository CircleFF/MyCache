package mycache

import (
	"MyCache/mycache/singleflight"
	"MyCache/mycache/utils/logger"
	"fmt"
	"log"
	"sync"
)

// Getter 缓存未命中时，从外部获取数据
type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

// Group 组，一个组中可以包含多个缓存节点
type Group struct {
	name       string
	mainCache  *cache
	getter     Getter
	nodeServer NodePicker
	loader     *singleflight.Group
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

// NewGroup 创建缓存组
func NewGroup(groupName string, strategyName string, nBytes int64, g Getter) *Group {
	if g == nil {
		panic("getter is nil")
	}

	mu.Lock()
	defer mu.Unlock()
	group := &Group{
		name:      groupName,
		mainCache: newCache(strategyName, nBytes),
		getter:    g,
		loader:    &singleflight.Group{},
	}

	groups[groupName] = group
	return group
}

func GetGroup(name string) *Group {
	mu.RLock()
	defer mu.RUnlock()
	return groups[name]
}

// RegisterNodes 将实现了 NodePicker 接口的 HTTPPool 注册到 Group 中
func (g *Group) RegisterNodes(nodes NodePicker) {
	if g.nodeServer != nil {
		panic("Register NodePicker called more than once")
	}
	g.nodeServer = nodes
}

// Get 获取缓存数据，如果缓存中没有，则从外部获取
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is empty")
	}

	if b, ok := g.mainCache.get(key); ok {
		log.Println("mycache hit")
		return b, nil
	}

	return g.load(key)
}

// load 加载数据，从其他节点或者本地获取
func (g *Group) load(key string) (ByteView, error) {
	bv, err := g.loader.Do(key, func() (interface{}, error) {
		// 再次从缓存中判断
		if val, ok := g.mainCache.get(key); ok {
			return val, nil
		}
		if g.nodeServer != nil {
			if node, ok := g.nodeServer.PickNode(key); ok {
				if value, err := g.loadFromNode(node, key); err == nil {
					return value, nil
				} else {
					log.Println("Failed to get from node", err)
				}
			}
		}
		return g.loadLocally(key)
	})

	if err != nil {
		return ByteView{}, err
	}
	return bv.(ByteView), nil
}

// loadLocally 从本地数据源 GetterFunc 加载数据，并添加到缓存
func (g *Group) loadLocally(key string) (ByteView, error) {
	val, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}

	b := ByteView{b: cloneBytes(val)}
	g.addData(key, b)
	return b, nil
}

func (g *Group) addData(key string, value ByteView) {
	g.mainCache.set(key, value)
}

// loadFromNode 从其他节点加载数据
func (g *Group) loadFromNode(node DataGetter, key string) (ByteView, error) {
	logger.Logger.Info("loading from other node")
	b, err := node.GetData(g.name, key)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: b}, nil
}
