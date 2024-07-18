package main

import (
	"MyCache/mycache"
	"MyCache/mycache/utils/logger"
	"flag"
	"fmt"
	"log"
)

var db = map[string]string{
	"k1": "v1",
	"k5": "v5",
	"k7": "v7",
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

func main() {
	var port int
	flag.IntVar(&port, "port", 8001, "mycache server port")
	flag.Parse()

	nodeAddrMap := map[int]string{
		8001: "localhost:8001",
		8002: "localhost:8002",
		8003: "localhost:8003",
	}

	nodeServiceNameMap := map[int]string{
		8001: "GroupCache",
		8002: "GroupCache",
		8003: "GroupCache",
	}

	var nodeAddrs []string
	var nodeServiceNames []string
	for k, v := range nodeAddrMap {
		nodeAddrs = append(nodeAddrs, v)
		nodeServiceNames = append(nodeServiceNames, nodeServiceNameMap[k])
	}

	g := createGroup()

	srv := mycache.NewGRPCSerer(nodeAddrMap[port])
	logger.Logger.Infoln("created a grpc server!")
	srv.Set(nodeAddrs, nodeServiceNames)
	g.RegisterNodes(srv)
	srv.Start("localhost:2379", nodeServiceNameMap[port])

}
