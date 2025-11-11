package solana

import (
	"context"
	"math/big"

	"github.com/reguluswee/walletus/common/chain/dep"
)

type SOLClient struct { /* rpc 池等 */
}

func NewSOLClient() *SOLClient { return &SOLClient{} }

func (c *SOLClient) Anchor(ctx context.Context, network string, cs dep.Consistency) (dep.AnchorRef, error) {
	return dep.AnchorRef{Height: 275019999, Tag: cs.Mode, Network: network, Provider: "sol-provider-A"}, nil
}

func (c *SOLClient) NativeBalance(ctx context.Context, network, address string, a dep.AnchorRef) (*dep.NativeBalance, error) {
	return &dep.NativeBalance{Symbol: "SOL", Decimals: 9, Amount: big.NewInt(5000000000)}, nil
}

func (c *SOLClient) TokenBalances(ctx context.Context, network, address string, tokens []string, a dep.AnchorRef) ([]dep.TokenBalance, error) {
	return nil, nil
}

func (c *SOLClient) NativeBalanceBatch(ctx context.Context, network string, addrs []string, a dep.AnchorRef) (map[string]*dep.NativeBalance, error) {
	out := map[string]*dep.NativeBalance{}
	for _, ad := range addrs {
		out[ad] = &dep.NativeBalance{Symbol: "SOL", Decimals: 9, Amount: big.NewInt(1)}
	}
	return out, nil
}

func (c *SOLClient) TokenBalancesBatch(ctx context.Context, network string, addr2tokens map[string][]string, a dep.AnchorRef) (map[string][]dep.TokenBalance, error) {
	return map[string][]dep.TokenBalance{}, nil
}

func MustRegister() {
	dep.Register(dep.GetSupportedSol(), NewSOLClient())
}
