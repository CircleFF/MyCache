package main

import (
	"MyCache/mycache"

	"flag"
	"fmt"
	"log"
	"net/http"
)

var db = map[string]string{
	"k1": "v1",
	"k2": "v2",
	"k3": "v3",
}

func createGroup() *mycache.Group {
	return mycache.NewGroup("test", "lru", 1<<10, mycache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
}

func startCacheServer(addr string, addrs []string, g *mycache.Group) {
	cacheServer := mycache.NewHTTPPool(addr)
	cacheServer.Set(addrs...)
	g.RegisterNodes(cacheServer)
	log.Println("mycache is running at", addr)
	log.Fatal(http.ListenAndServe(addr[7:], cacheServer))
}

func startAPIServer(addr string, g *mycache.Group) {
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			bv, err := g.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(bv.ByteSlice())
		}))
	log.Println("frontend server is running at", addr)
	log.Fatal(http.ListenAndServe(addr[7:], nil))

}

func main() {
	var port int
	var api bool
	flag.IntVar(&port, "port", 8001, "mycache server port")
	flag.BoolVar(&api, "api", false, "if start a api server")
	flag.Parse()

	apiAddr := "http://localhost:8888"
	nodeAddrMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}

	var nodeAddrs []string
	for _, v := range nodeAddrMap {
		nodeAddrs = append(nodeAddrs, v)
	}

	g := createGroup()
	if api {
		go startAPIServer(apiAddr, g)
	}
	startCacheServer(nodeAddrMap[port], nodeAddrs, g)
}
