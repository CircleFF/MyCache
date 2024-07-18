package main

import (
	"MyCache/mycache/etcd"
	pb "MyCache/mycache/mycachepb"
	"MyCache/mycache/utils/logger"
	"time"

	"golang.org/x/net/context"
)

var db = []string{
	"k1", "k4", "k5", "k6", "k7",
}

func main() {
	conn, err := etcd.Discovery("localhost:2379", "GroupCache")
	if err != nil {
		logger.Logger.Errorf("start client error:%s", err.Error())
		return
	}

	client := pb.NewGroupCacheClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	for _, k := range db {
		ctxv := context.WithValue(ctx, "key", k)
		resp, err := client.Get(ctxv, &pb.Request{
			Group: "test",
			Key:   k,
		})
		if err != nil {
			logger.Logger.Errorf("client get key:%s from group:%s error:%s", k, "test", err.Error())
			continue
		}
		logger.Logger.Infof("client get key:%s success, value=%s", k, string(resp.Value))
	}
	time.Sleep(60 * time.Second)
}
