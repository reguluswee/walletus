package chain

import (
	"context"
	"time"

	"github.com/reguluswee/walletus/common/chain/dep"
	"github.com/reguluswee/walletus/common/chain/evm"
)

type Gateway struct{}

func NewGateway() *Gateway { return &Gateway{} }

type BalanceQuery struct {
	Chain       dep.ChainDef
	Network     string
	Addresses   []string
	Tokens      map[string][]string
	Consistency dep.Consistency
}

func init() {
	evm.MustRegister()
}

func (g *Gateway) GetBalances(ctx context.Context, q BalanceQuery) (*dep.BatchBalanceResult, error) {
	client, ok := dep.GetClient(q.Chain)

	if !ok {
		return nil, dep.ErrUnsupportedChain
	}

	anchor, err := client.Anchor(ctx, q.Chain.Name, q.Consistency)
	if err != nil {
		return nil, err
	}

	nativeMap, err := client.NativeBalanceBatch(ctx, q.Chain.Name, q.Addresses, anchor)
	if err != nil {
		return nil, err
	}

	var tokenMap map[string][]dep.TokenBalance
	if len(q.Tokens) > 0 {
		tokenMap, err = client.TokenBalancesBatch(ctx, q.Chain.Name, q.Tokens, anchor)
		if err != nil {
			return nil, err
		}
	} else {
		tokenMap = map[string][]dep.TokenBalance{}
	}
	out := &dep.BatchBalanceResult{Chain: q.Chain}
	now := time.Now().UTC().Format(time.RFC3339)
	for _, addr := range q.Addresses {
		br := dep.BalanceResult{
			Anchor:       anchor,
			Address:      addr,
			Native:       nativeMap[addr],
			Tokens:       tokenMap[addr],
			QueriedAtUTC: now,
		}
		out.Results = append(out.Results, br)
	}
	return out, nil
}
