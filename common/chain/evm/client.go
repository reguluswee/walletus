package evm

import (
	"context"
	"math/big"

	"github.com/reguluswee/walletus/common/chain/dep"
)

type EVMClient struct {
	// 放 RPC provider 池、multicall 合约地址、单飞/批处理器等
}

func NewEVMClient( /*cfg*/ ) *EVMClient { return &EVMClient{} }

func (c *EVMClient) Anchor(ctx context.Context, network string, cs dep.Consistency) (dep.AnchorRef, error) {
	// 例：根据 cs.Mode=latest|safe|finalized 取块高；返回 AnchorRef
	// blockNumber, provider := rpc.GetBlockNumber(tag)
	return dep.AnchorRef{
		Height:   20987654,
		Tag:      cs.Mode,
		Network:  network,
		Provider: "evm-provider-A",
	}, nil
}

func (c *EVMClient) NativeBalance(ctx context.Context, network, address string, a dep.AnchorRef) (*dep.NativeBalance, error) {
	// 直接 eth_getBalance(address, a.Height/tag)
	return &dep.NativeBalance{
		Symbol: "ETH", Decimals: 18, Amount: big.NewInt(0).SetUint64(1234567890),
	}, nil
}

func (c *EVMClient) TokenBalances(ctx context.Context, network, address string, tokens []string, a dep.AnchorRef) ([]dep.TokenBalance, error) {
	// 使用 Multicall 合并 balanceOf(address)；这里演示返回空
	return nil, nil
}

func (c *EVMClient) NativeBalanceBatch(ctx context.Context, network string, addrs []string, a dep.AnchorRef) (map[string]*dep.NativeBalance, error) {
	// 批量 eth_getBalance（或 multicall）
	out := make(map[string]*dep.NativeBalance, len(addrs))
	for _, ad := range addrs {
		out[ad] = &dep.NativeBalance{Symbol: "ETH", Decimals: 18, Amount: big.NewInt(1)}
	}
	return out, nil
}

func (c *EVMClient) TokenBalancesBatch(ctx context.Context, network string, addr2tokens map[string][]string, a dep.AnchorRef) (map[string][]dep.TokenBalance, error) {
	// 将 (addr, tokens[]) 拆成批次，用 Multicall 一次性查
	return map[string][]dep.TokenBalance{}, nil
}

// 注册（在 main 或 init 中）
func MustRegister() {
	for _, v := range dep.GetSupportedEVMs() {
		dep.Register(v, NewEVMClient())
	}
}
