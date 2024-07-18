package mycache

import (
	"MyCache/mycache/consistent_hash"
	"MyCache/mycache/utils/logger"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

var (
	defaultBasePath = "/mycache/"
	defaultReplicas = 3
)

var _ NodePicker = (*HTTPPool)(nil)

type HTTPPool struct {
	// selfAddr host:port
	selfAddr string
	basePath string

	mu        sync.Mutex
	hash      *consistent_hash.Hash
	getterMap map[string]*httpGetter
}

func NewHTTPPool(selfAddr string) *HTTPPool {
	return &HTTPPool{
		selfAddr: selfAddr,
		basePath: defaultBasePath,
	}
}

func (h *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[server]: %s, %s", h.selfAddr, fmt.Sprintf(format, v...))
}

// ServeHTTP 实现 http.Handler 接口，用于处理 http 请求
func (h *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("method:%s, path:%s", r.Method, r.URL.Path)

	if !strings.HasPrefix(r.URL.Path, h.basePath) {
		logger.Logger.Errorf("unexpected path %s", r.URL.Path)
	}
	paths := strings.SplitN(r.URL.Path[len(h.basePath):], "/", 2)
	if len(paths) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	groupName := paths[0]
	key := paths[1]

	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "group not exist", http.StatusNotFound)
		return
	}
	val, err := group.Get(key)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/octet-stream")
	w.Write(val.ByteSlice())
}

// Set 设置其他节点
func (h *HTTPPool) Set(nodes ...string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.hash = consistent_hash.NewHash(defaultReplicas, nil)
	h.hash.AddNode(nodes...)
	h.getterMap = make(map[string]*httpGetter, len(nodes))
	for _, node := range nodes {
		h.getterMap[node] = &httpGetter{baseURL: node + h.basePath}
	}

}

// PickNode 通过 key 来获取节点
func (h *HTTPPool) PickNode(key string) (DataGetter, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()

	node := h.hash.GetNode(key)
	if node != "" && node != h.selfAddr {
		h.Log("pick node: %s", node)
		return h.getterMap[node], true
	}

	return nil, false
}

// httpGetter 实现 DataGetter，用于获取其他节点数据
type httpGetter struct {
	baseURL string
}

var _ DataGetter = (*httpGetter)(nil)

// GetData 获取其他节点数据
func (h *httpGetter) GetData(group, key string) ([]byte, error) {
	u := fmt.Sprintf("%v%v/%v", h.baseURL, url.QueryEscape(group), url.QueryEscape(key))

	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed status code is %v", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body failed, %v", err)
	}

	return data, nil
}
