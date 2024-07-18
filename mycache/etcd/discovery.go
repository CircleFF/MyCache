package etcd

import (
	"MyCache/mycache/balancer"
	"MyCache/mycache/utils/logger"
	"fmt"
	"strings"
	"time"

	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
)

func Discovery(etcdAddr, serviceName string) (*grpc.ClientConn, error) {
	// 注册 etcd 解析器
	r := newResolver(etcdAddr)
	resolver.Register(r)

	// 客户端连接服务器会同步调用 r.Build()
	conn, err := grpc.NewClient(r.Scheme()+"://author/"+serviceName, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"LoadBalancingPolicy": "%s"}`, balancer.Consistent_Name_LB)))
	logger.Logger.Infof("target = %s", conn.Target())
	if err != nil {
		logger.Logger.Errorf("connect to etcd server error:%s", err.Error())
		return nil, err
	}
	return conn, nil
}

// EtcdResolver etcd 解析器
type EtcdResolver struct {
	EtcdAddr   string
	ClientConn resolver.ClientConn
}

func newResolver(etcdAddr string) resolver.Builder {
	return &EtcdResolver{EtcdAddr: etcdAddr}
}

func (r *EtcdResolver) Scheme() string {
	return scheme
}

// ResolveNow watch 有变化以后会调用
func (r *EtcdResolver) ResolveNow(rn resolver.ResolveNowOptions) {
	logger.Logger.Infoln("etcd Resolve Now")
}

// Close 解析器关闭时调用
func (r *EtcdResolver) Close() {
	logger.Logger.Infoln("etcd close")
}

// Build 构建解析器
func (r *EtcdResolver) Build(target resolver.Target, clientConn resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	logger.Logger.Info("resolver is building")
	var err error

	// 构建 etcd client
	if cli == nil {
		cli, err = clientv3.New(clientv3.Config{
			Endpoints:   strings.Split(r.EtcdAddr, ";"),
			DialTimeout: 15 * time.Second,
		})
		if err != nil {
			logger.Logger.Errorf("build etcd client error:%s", err.Error())
			return nil, err
		}
	}

	r.ClientConn = clientConn

	// 监听的功能单独出来
	go r.Watch("/" + scheme + "/" + target.Endpoint() + "/")

	return r, nil
}

// Watch 监听 etcd 中某个 key 前缀的服务地址列表的变化
func (r *EtcdResolver) Watch(keyPrefix string) {
	logger.Logger.Infof("resolver is watching, key perfix is %s", keyPrefix)
	var addrList []resolver.Address

	resp, err := cli.Get(context.Background(), keyPrefix, clientv3.WithPrefix())
	logger.Logger.Info("get kvs from etcd")
	if err != nil {
		logger.Logger.Errorf("get server address list error:%s", err.Error())
	} else {
		for i := range resp.Kvs {
			addrList = append(addrList, resolver.Address{Addr: string(resp.Kvs[i].Value)})
		}
	}
	logger.Logger.Infof("addrList = %v", addrList)

	r.ClientConn.UpdateState(resolver.State{Addresses: addrList})

	// 监听服务地址列表的变化
	rch := cli.Watch(context.Background(), keyPrefix, clientv3.WithPrefix())
	for n := range rch {
		for _, ev := range n.Events {
			addr := strings.TrimPrefix(string(ev.Kv.Key), keyPrefix)
			switch ev.Type {
			case mvccpb.PUT:
				if !exists(addrList, addr) {
					logger.Logger.Infof("watched put addr:%s", addr)
					addrList = append(addrList, resolver.Address{Addr: addr})
					r.ClientConn.UpdateState(resolver.State{Addresses: addrList})
				}
			case mvccpb.DELETE:
				if s, ok := remove(addrList, addr); ok {
					logger.Logger.Infof("watched delete addr:%s", addr)
					addrList = s
					r.ClientConn.UpdateState(resolver.State{Addresses: addrList})
				}
			}
		}
	}
}

func exists(l []resolver.Address, addr string) bool {
	for i := range l {
		if l[i].Addr == addr {
			return true
		}
	}
	return false
}

func remove(s []resolver.Address, addr string) ([]resolver.Address, bool) {
	for i := range s {
		if s[i].Addr == addr {
			s[i] = s[len(s)-1]
			return s[:len(s)-1], true
		}
	}
	return nil, false
}
