package chain

import (
	"context"
	"testing"

	"github.com/reguluswee/walletus/common/chain/dep"
	"github.com/reguluswee/walletus/common/chain/evm"
	"github.com/reguluswee/walletus/common/chain/solana"
	"github.com/reguluswee/walletus/common/chain/tron"
)

func TestChainGateway(t *testing.T) {
	evm.MustRegister()
	solana.MustRegister()
	tron.MustRegister()

	eth := dep.ChainDef{
		Name:     "ETH",
		CoinType: 60,
	}
	gw := NewGateway()
	q := BalanceQuery{
		Chain:       eth,
		Network:     "mainnet",
		Addresses:   []string{"0xA...", "0xB..."},
		Tokens:      map[string][]string{ /* 可空：只查原生 */ },
		Consistency: dep.Consistency{Mode: "safe", MinConfirmations: 0},
	}
	res, err := gw.GetBalances(context.Background(), q)
	if err != nil {
		panic(err)
	}
	t.Log(res)
}
