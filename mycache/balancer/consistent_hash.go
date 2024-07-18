package balancer

import (
	"MyCache/mycache/consistent_hash"
	"MyCache/mycache/utils/logger"

	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
)

const Consistent_Name_LB = "consistent_hash"

func init() {
	balancer.Register(base.NewBalancerBuilder(Consistent_Name_LB, &chPickerBuilder{}, base.Config{HealthCheck: true}))
}

type chPickerBuilder struct {
}

func (b *chPickerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	if len(info.ReadySCs) == 0 {
		return base.NewErrPicker(balancer.ErrNoSubConnAvailable)
	}

	picker := &chPicker{
		subConns: make(map[string]balancer.SubConn),
		hash:     consistent_hash.NewHash(3, nil),
	}
	for conn, connInfo := range info.ReadySCs {
		node := connInfo.Address.Addr
		picker.hash.AddNode(node)
		picker.subConns[node] = conn
	}
	logger.Logger.Infoln("build hash ring completed!")
	return picker

}

type chPicker struct {
	subConns map[string]balancer.SubConn
	hash     *consistent_hash.Hash
}

func (p *chPicker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	ret := balancer.PickResult{}
	logger.Logger.Infoln(p.hash)
	logger.Logger.Infof("pick ctx = %v", info.Ctx)
	key := info.Ctx.Value("key").(string)
	addr := p.hash.GetNode(key)
	ret.SubConn = p.subConns[addr]
	logger.Logger.Infof("balancer pick addr %s", addr)
	return ret, nil
}
