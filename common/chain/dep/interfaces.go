package dep

import "context"

type Reader interface {
	// Anchor：根据一致性策略取“读锚点”，保证一次请求内的一致性
	Anchor(ctx context.Context, network string, c Consistency) (AnchorRef, error)

	// Single Query
	NativeBalance(ctx context.Context, network string, address string, anchor AnchorRef) (*NativeBalance, error)
	TokenBalances(ctx context.Context, network string, address string, tokens []string, anchor AnchorRef) ([]TokenBalance, error)

	// Batch Query
	NativeBalanceBatch(ctx context.Context, network string, addresses []string, anchor AnchorRef) (map[string]*NativeBalance, error)
	TokenBalancesBatch(ctx context.Context, network string, addressToTokens map[string][]string, anchor AnchorRef) (map[string][]TokenBalance, error)
}

type Signer interface {
	SignTransferNative(ctx context.Context, network string, from, to string, amount string, opts any) (rawTx []byte, txHash string, err error)
	SignTransferToken(ctx context.Context, network, token, from, to string, amount string, opts any) (rawTx []byte, txHash string, err error)
}

type Boradcaster interface {
	Boradcast(ctx context.Context, network string, rawTx []byte)
}

type Sweeper interface {
	BuildAndSignSweep(ctx context.Context, network string, from string, dest string, opts any) (rawTx []byte, txHash string, err error)
}

type Client interface {
	Reader
}
