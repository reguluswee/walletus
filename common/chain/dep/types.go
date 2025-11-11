package dep

import (
	"fmt"
	"math/big"
)

type ChainCode string

type ChainDef struct {
	Name     string
	CoinType uint32
}

var chains = []ChainDef{
	{"ETH", 60},
	{"BSC", 60},
	{"OP", 60},
	{"ARB", 60},
	{"POLYGON", 60},
	{"TRON", 195},
	{"SOLANA", 501},
}

var supportedEVMs = []ChainDef{
	{"ETH", 60},
	{"BSC", 60},
	{"OP", 60},
	{"ARB", 60},
	{"POLYGON", 60},
}

var supportedSol = ChainDef{
	Name:     "SOLANA",
	CoinType: 501,
}

var supportedTron = ChainDef{
	Name:     "TRON",
	CoinType: 195,
}

func GetSupportedEVMs() []ChainDef {
	return supportedEVMs
}

func GetSupportedSol() ChainDef {
	return supportedSol
}

func GetSupportedTron() ChainDef {
	return supportedTron
}

func GetSupportedChains() []ChainDef {
	return chains
}

type Consistency struct {
	// EVM: latest|safe|finalized
	// Solana: processed|confirmed|finalized
	// Tron: latest|latest_solid
	Mode             string
	MinConfirmations int
}

type AnchorRef struct {
	// EVM: BlockNumber；Solana: Slot；Tron: BlockNumber
	Height   uint64
	Tag      string
	Network  string
	Provider string
}

type NativeBalance struct {
	Symbol   string
	Decimals int
	Amount   *big.Int
}

type TokenBalance struct {
	Contract string
	NativeBalance
}

type BalanceResult struct {
	Anchor       AnchorRef
	Address      string
	Native       *NativeBalance
	Tokens       []TokenBalance
	QueriedAtUTC string
}

type BatchBalanceResult struct {
	Chain   ChainDef
	Results []BalanceResult
}

var (
	ErrUnsupportedChain = fmt.Errorf("unsupported chain code")
	ErrInvalidAddress   = fmt.Errorf("invalid address")
	ErrRPCFailed        = fmt.Errorf("rpc failed")
)
