package tron

import (
	"context"
	"math/big"

	"github.com/reguluswee/walletus/common/chain/dep"
)

type TRXClient struct {
}

func NewTRXClient() *TRXClient { return &TRXClient{} }

func (c *TRXClient) Anchor(ctx context.Context, network string, cs dep.Consistency) (dep.AnchorRef, error) {
	return dep.AnchorRef{Height: 51234567, Tag: cs.Mode, Network: network, Provider: "tron-provider-A"}, nil
}
func (c *TRXClient) NativeBalance(ctx context.Context, network, address string, a dep.AnchorRef) (*dep.NativeBalance, error) {
	return &dep.NativeBalance{Symbol: "TRX", Decimals: 6, Amount: big.NewInt(123)}, nil
}
func (c *TRXClient) TokenBalances(ctx context.Context, network, address string, tokens []string, a dep.AnchorRef) ([]dep.TokenBalance, error) {
	return nil, nil
}
func (c *TRXClient) NativeBalanceBatch(ctx context.Context, network string, addrs []string, a dep.AnchorRef) (map[string]*dep.NativeBalance, error) {
	out := map[string]*dep.NativeBalance{}
	for _, ad := range addrs {
		out[ad] = &dep.NativeBalance{Symbol: "TRX", Decimals: 6, Amount: big.NewInt(1)}
	}
	return out, nil
}
func (c *TRXClient) TokenBalancesBatch(ctx context.Context, network string, addr2tokens map[string][]string, a dep.AnchorRef) (map[string][]dep.TokenBalance, error) {
	return map[string][]dep.TokenBalance{}, nil
}
func MustRegister() {
	dep.Register(dep.GetSupportedTron(), NewTRXClient())
}
