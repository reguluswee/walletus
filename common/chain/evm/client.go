package evm

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	gethrpc "github.com/ethereum/go-ethereum/rpc"
	"github.com/reguluswee/walletus/common/chain/dep"
	"golang.org/x/sync/singleflight"
)

type rpcPool struct {
	clients []*gethrpc.Client
	names   []string
}

type EVMClient struct {
	mu           sync.RWMutex
	pools        map[string]*rpcPool // network -> pool
	mcAddr       map[string]string   // network -> multicall addr
	sf           singleflight.Group
	maxBatch     int
	reqTimeout   time.Duration
	erc20ABI     abi.ABI
	multicallABI abi.ABI
}

func NewEVMClient(chain dep.ChainDef) *EVMClient {
	iface := &EVMClient{
		pools:      make(map[string]*rpcPool),
		mcAddr:     make(map[string]string),
		maxBatch:   256,
		reqTimeout: 2 * time.Second,
	}
	erc20, _ := abi.JSON(strings.NewReader(erc20ABIJSON))
	mc, _ := abi.JSON(strings.NewReader(multicallABI))
	iface.erc20ABI = erc20
	iface.multicallABI = mc
	return iface
}

func (c *EVMClient) Anchor(ctx context.Context, network string, cs dep.Consistency) (dep.AnchorRef, error) {
	tag := cs.Mode
	if tag == "" {
		tag = "latest"
	}
	rc, name, err := c.pick(network)
	if err != nil {
		return dep.AnchorRef{}, err
	}

	ctx2, cancel := context.WithTimeout(ctx, c.reqTimeout)
	defer cancel()

	var hexBlock string
	switch tag {
	case "latest", "safe", "finalized":
		var head struct {
			Number string `json:"number"`
		}
		if err := rc.CallContext(ctx2, &head, "eth_getBlockByNumber", tag, false); err != nil {
			return dep.AnchorRef{}, err
		}
		if head.Number == "" {
			return dep.AnchorRef{}, fmt.Errorf("empty header for tag=%s", tag)
		}
		hexBlock = head.Number
	default:
		// 十进制或0x
		if strings.HasPrefix(tag, "0x") {
			hexBlock = tag
		} else {
			n := new(big.Int)
			if _, ok := n.SetString(tag, 10); !ok {
				return dep.AnchorRef{}, fmt.Errorf("invalid block tag: %s", tag)
			}
			hexBlock = "0x" + n.Text(16)
		}
	}

	height := hexToUint64(hexBlock)
	return dep.AnchorRef{
		Height:   height,
		Tag:      cs.Mode,
		Network:  network,
		Provider: name,
	}, nil
}

func (c *EVMClient) NativeBalance(ctx context.Context, network, address string, a dep.AnchorRef) (*dep.NativeBalance, error) {
	mp, err := c.NativeBalanceBatch(ctx, network, []string{address}, a)
	if err != nil {
		return nil, err
	}
	return mp[address], nil
}

func (c *EVMClient) TokenBalances(ctx context.Context, network, address string, tokens []string, a dep.AnchorRef) ([]dep.TokenBalance, error) {
	mp, err := c.TokenBalancesBatch(ctx, network, map[string][]string{address: tokens}, a)
	if err != nil {
		return nil, err
	}
	return mp[address], nil
}

func (c *EVMClient) NativeBalanceBatch(ctx context.Context, network string, addrs []string, a dep.AnchorRef) (map[string]*dep.NativeBalance, error) {
	if len(addrs) == 0 {
		return map[string]*dep.NativeBalance{}, nil
	}
	key := fmt.Sprintf("nb:%s:%d:%s", network, a.Height, hashStrings(addrs))
	v, err, _ := c.sf.Do(key, func() (interface{}, error) {
		rc, _, err := c.pick(network)
		if err != nil {
			return nil, err
		}
		out := make(map[string]*dep.NativeBalance, len(addrs))
		for _, batch := range chunkStrings(addrs, c.maxBatch) {
			results := make([]string, len(batch))
			elems := make([]gethrpc.BatchElem, 0, len(batch))
			for i, ad := range batch {
				elems = append(elems, gethrpc.BatchElem{
					Method: "eth_getBalance",
					Args:   []any{common.HexToAddress(ad), blockParam(a)},
					Result: &results[i],
				})
			}
			ctx2, cancel := context.WithTimeout(ctx, c.reqTimeout)
			err := rc.BatchCallContext(ctx2, elems)
			cancel()
			if err != nil {
				return nil, err
			}
			for i, ad := range batch {
				out[ad] = &dep.NativeBalance{
					Symbol:   nativeSymbol(network),
					Decimals: nativeDecimals(network),
					Amount:   hexToBig(results[i]),
				}
			}
		}
		return out, nil
	})
	if err != nil {
		return nil, err
	}
	return v.(map[string]*dep.NativeBalance), nil
}

func (c *EVMClient) TokenBalancesBatch(ctx context.Context, network string, addr2tokens map[string][]string, a dep.AnchorRef) (map[string][]dep.TokenBalance, error) {
	pairs := flattenPairs(addr2tokens) // [token, owner]
	if len(pairs) == 0 {
		return map[string][]dep.TokenBalance{}, nil
	}
	key := fmt.Sprintf("tb:%s:%d:%s", network, a.Height, hashPairs(pairs))
	v, err, _ := c.sf.Do(key, func() (interface{}, error) {
		// 先尝试 Multicall
		if mc := c.getMulticallAddr(network); mc != "" {
			if res, e := c.multicallBalanceOf(ctx, network, pairs, a); e == nil {
				return explodePairs(res), nil
			}
			// 失败再降级
		}
		res, e := c.batchedBalanceOf(ctx, network, pairs, a)
		if e != nil {
			return nil, e
		}
		return explodePairs(res), nil
	})
	if err != nil {
		return nil, err
	}
	return v.(map[string][]dep.TokenBalance), nil
}

// 注册（在 main 或 init 中）
func MustRegister() {
	for _, v := range dep.GetSupportedEVMs() {
		dep.Register(v, NewEVMClient(v))
	}
}
