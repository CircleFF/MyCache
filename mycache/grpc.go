package mycache

import (
	"MyCache/mycache/consistent_hash"
	"MyCache/mycache/etcd"
	pb "MyCache/mycache/mycachepb"
	"MyCache/mycache/utils/logger"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"google.golang.org/grpc"
)

var _ pb.GroupCacheServer = (*GRPCServer)(nil)
var _ NodePicker = (*GRPCServer)(nil)
var _ DataGetter = (*GRPCClient)(nil)

var etcdAddress string

type GRPCServer struct {
	pb.UnimplementedGroupCacheServer

	selfAddr  string
	mu        sync.Mutex
	hash      *consistent_hash.Hash
	clientMap map[string]*GRPCClient
}

type Entry struct {
	Addr        string
	ServiceName string
}

func NewGRPCSerer(addr string) *GRPCServer {
	return &GRPCServer{
		selfAddr: addr,
	}
}

func (g *GRPCServer) Set(nodes, names []string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.hash = consistent_hash.NewHash(defaultReplicas, nil)
	g.hash.AddNode(nodes...)
	g.clientMap = make(map[string]*GRPCClient)
	for i, node := range nodes {
		g.clientMap[node] = &GRPCClient{serviceName: names[i]}
	}

}

// PickNode 根据 key 获取节点
func (g *GRPCServer) PickNode(key string) (DataGetter, bool) {
	g.mu.Lock()
	defer g.mu.Unlock()

	logger.Logger.Info("is picking node")
	if node := g.hash.GetNode(key); node != "" && node != g.selfAddr {
		logger.Logger.Infof("selfAddr:%s, pick node addr:%s", g.selfAddr, node)
		return g.clientMap[node], true
	}
	logger.Logger.Info("need load locally")
	return nil, false
}

func (g *GRPCServer) Get(ctx context.Context, request *pb.Request) (*pb.Response, error) {
	groupName, key := request.Group, request.Key
	group := groups[groupName]
	if group == nil {
		return nil, fmt.Errorf("group:%s not found", groupName)
	}

	bv, err := group.Get(key)
	if err != nil {
		return nil, err
	}

	resp := &pb.Response{Value: bv.ByteSlice()}
	return resp, nil

}

func (g *GRPCServer) Start(etcdAddr, serviceName string) {
	if etcdAddr == "" {
		logger.Logger.Panicln("etcd address is empty")
	}
	etcdAddress = etcdAddr
	port := strings.Split(g.selfAddr, ":")[1]
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Printf("listen network failed, err:%v", err)
		return
	}
	defer listener.Close()

	srv := grpc.NewServer()
	defer srv.GracefulStop()

	// GRPCServer 结构体注册到 grpc 服务中
	pb.RegisterGroupCacheServer(srv, g)

	// 服务注册到 etcd 中
	etcd.Register(etcdAddr, serviceName, g.selfAddr, 6)
	logger.Logger.Infof("register %s to etcd %s", g.selfAddr, etcdAddr)

	// 启动一个客户端监听 etcd，更新处理

	// 关闭信号处理
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP, syscall.SIGQUIT)
	go func() {
		s := <-ch
		etcd.UnRegister("GroupCache", g.selfAddr)
		if i, ok := s.(syscall.Signal); ok {
			os.Exit(int(i))
		} else {
			os.Exit(0)
		}
	}()

	// 监听
	err = srv.Serve(listener)
	if err != nil {
		log.Printf("listen err:%v", err)
		return
	}

}

type GRPCClient struct {
	serviceName string
}

func (g *GRPCClient) GetData(group, key string) ([]byte, error) {
	conn, err := etcd.Discovery(etcdAddress, g.serviceName)
	logger.Logger.Infof("etcd address:%s, serviceName:%s", etcdAddress, g.serviceName)
	if err != nil {
		logger.Logger.Errorf("get etcd client error:%s", err.Error())
		return nil, err
	}
	logger.Logger.Info("other node connect to etcd...")
	defer conn.Close()

	cli := pb.NewGroupCacheClient(conn)
	logger.Logger.Info("create a client")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	ctxv := context.WithValue(ctx, "mycache", key)
	logger.Logger.Infof("grpc ctxv = %v", ctxv)
	resp, err := cli.Get(ctxv, &pb.Request{
		Group: group,
		Key:   key,
	})
	if err != nil {
		logger.Logger.Errorf("grpc client get data error:%s", err.Error())
		return nil, err
	}
	return resp.Value, nil
}
