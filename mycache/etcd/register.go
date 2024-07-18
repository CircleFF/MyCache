package etcd

import (
	"MyCache/mycache/utils/logger"
	"strings"
	"time"

	"golang.org/x/net/context"

	clientv3 "go.etcd.io/etcd/client/v3"
)

const scheme = "etcd"

var cli *clientv3.Client

// Register 将服务注册到 etcd 中
func Register(etcdAddr, serviceName, serverAddr string, ttl int64) error {
	var err error

	if cli == nil {
		// 构建 etcd client
		cli, err = clientv3.New(clientv3.Config{
			Endpoints:   strings.Split(etcdAddr, ";"),
			DialTimeout: 15 * time.Second,
		})
		if err != nil {
			logger.Logger.Errorf("build etcd client error:%s", err.Error())
			return err
		}
	}

	// 与 etcd 建立长连接，并保证连接不断(心跳检测)
	ticker := time.NewTicker(time.Second * time.Duration(ttl))
	go func() {
		key := "/" + scheme + "/" + serviceName + "/" + serverAddr
		logger.Logger.Infof("key=%s", key)
		for {
			resp, err := cli.Get(context.Background(), key)
			//fmt.Printf("resp:%+v\n", resp)
			if err != nil {
				logger.Logger.Errorf("get etcd server error:%s", err.Error())
			} else if resp.Count == 0 { //尚未注册
				err = keepAlive(serviceName, serverAddr, ttl)
				if err != nil {
					logger.Logger.Errorf("keep alive error:%s", err.Error())
				}
				logger.Logger.Info("regiseter to etcd success")
			}
			<-ticker.C
		}
	}()

	return nil
}

// keepAlive 保持服务器与 etcd 的长连接
func keepAlive(serviceName, serverAddr string, ttl int64) error {
	// 创建租约
	leaseResp, err := cli.Grant(context.Background(), ttl)
	if err != nil {
		logger.Logger.Errorf("build lease error:%s", err.Error())
		return err
	}

	// 将服务地址注册到 etcd 中
	key := "/" + scheme + "/" + serviceName + "/" + serverAddr
	_, err = cli.Put(context.Background(), key, serverAddr, clientv3.WithLease(leaseResp.ID))
	if err != nil {
		logger.Logger.Errorf("register service to etcd error:%s", err.Error())
		return err
	}

	// 建立长连接
	ch, err := cli.KeepAlive(context.Background(), leaseResp.ID)
	if err != nil {
		logger.Logger.Errorf("build keep alive error:%s", err.Error())
		return err
	}
	logger.Logger.Info("build long connection success")

	// 清空 KeepAlive 返回的 channel
	go func() {
		for {
			<-ch
		}
	}()
	return nil
}

// UnRegister 取消注册
func UnRegister(serviceName, serverAddr string) {
	if cli != nil {
		key := "/" + scheme + "/" + serviceName + "/" + serverAddr
		cli.Delete(context.Background(), key)
		logger.Logger.Infof("etcd unregister key:%s", key)
	}
}
