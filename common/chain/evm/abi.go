package evm

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	gethrpc "github.com/ethereum/go-ethereum/rpc"
	"github.com/reguluswee/walletus/common/chain/dep"
	"github.com/reguluswee/walletus/common/config"
	"golang.org/x/crypto/sha3"
)

func (c *EVMClient) pick(network string) (*gethrpc.Client, string, error) {
	pool, err := c.ensurePool(network)
	if err != nil {
		return nil, "", err
	}
	// 简单：取第一个（可扩展为加权/熔断/健康检查）
	if len(pool.clients) == 0 {
		return nil, "", fmt.Errorf("no RPC available for %s", network)
	}
	return pool.clients[0], pool.names[0], nil
}

// ensurePool 确保指定网络的 RPC 连接池已初始化
// 该方法会从系统配置中读取 RPC 端点配置（通过 config.GetRpcConfig）
// 配置文件的路径由 config 包管理，通常从 dev.yml 或环境变量指定
// RPC 配置格式：chain[].name 必须与 network 参数匹配，chain[].queryRpc 包含 RPC 端点列表
func (c *EVMClient) ensurePool(network string) (*rpcPool, error) {
	c.mu.RLock()
	p, ok := c.pools[network]
	c.mu.RUnlock()
	if ok {
		return p, nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	// 双检锁，避免并发创建
	if p, ok := c.pools[network]; ok {
		return p, nil
	}

	// 从系统配置中获取 RPC 配置
	// 配置通过 config.GetRpcConfig 读取，匹配链名称（如 "ETH", "BSC", "BSC_TESTNET"）
	cc := config.GetRpcConfig(network)
	if cc == nil || len(cc.GetRpc()) == 0 {
		return nil, fmt.Errorf("missing RPC config for %s", network)
	}

	// 创建 RPC 连接池
	var pool rpcPool
	for _, url := range cc.GetRpc() {
		if rc, err := gethrpc.Dial(url); err == nil {
			pool.clients = append(pool.clients, rc)
			pool.names = append(pool.names, shortAlias(network, url))
		}
		// 注意：如果某个 RPC 连接失败，会静默跳过，只使用成功连接的 RPC
	}
	if len(pool.clients) == 0 {
		return nil, fmt.Errorf("dial RPC failed for %s", network)
	}
	c.pools[network] = &pool

	// 初始化 Multicall 地址（如果尚未初始化）
	if _, ok := c.mcAddr[network]; !ok {
		c.mcAddr[network] = resolveMulticall(network)
	}
	return &pool, nil
}

func (c *EVMClient) getMulticallAddr(network string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.mcAddr[network]
}

// ---------------- 工具函数 ----------------

func shortAlias(chain, url string) string {
	u := url
	if i := strings.Index(u, "://"); i >= 0 {
		u = u[i+3:]
	}
	if j := strings.Index(u, "/"); j >= 0 {
		u = u[:j]
	}
	if len(u) > 28 {
		u = u[:28]
	}
	return strings.ToLower(chain) + ":" + u
}

func nativeSymbol(network string) string {
	switch strings.ToLower(network) {
	case "bsc", "bnb", "bnbchain":
		return "BNB"
	case "bsc_testnet":
		return "tBNB"
	case "polygon", "matic":
		return "MATIC"
	default:
		return "ETH"
	}
}
func nativeDecimals(string) int { return 18 }

func blockParam(a dep.AnchorRef) any {
	switch a.Tag {
	case "latest", "safe", "finalized":
		return a.Tag
	}
	if a.Height == 0 {
		return "latest"
	}
	return fmt.Sprintf("0x%x", a.Height)
}

func hexToBig(s string) *big.Int {
	if s == "" {
		return big.NewInt(0)
	}
	n := new(big.Int)
	n.SetString(strings.TrimPrefix(s, "0x"), 16)
	return n
}

func hexToUint64(s string) uint64 {
	if s == "" {
		return 0
	}
	n := new(big.Int)
	n.SetString(strings.TrimPrefix(s, "0x"), 16)
	return n.Uint64()
}

func chunkStrings(ss []string, n int) [][]string {
	if n <= 0 || len(ss) <= n {
		return [][]string{ss}
	}
	var out [][]string
	for i := 0; i < len(ss); i += n {
		j := i + n
		if j > len(ss) {
			j = len(ss)
		}
		out = append(out, ss[i:j])
	}
	return out
}

func chunkPairs(pairs [][2]common.Address, n int) [][][2]common.Address {
	if n <= 0 || len(pairs) <= n {
		return [][][2]common.Address{pairs}
	}
	var out [][][2]common.Address
	for i := 0; i < len(pairs); i += n {
		j := i + n
		if j > len(pairs) {
			j = len(pairs)
		}
		out = append(out, pairs[i:j])
	}
	return out
}

// addr->tokens 展平为 [token, owner]
func flattenPairs(m map[string][]string) [][2]common.Address {
	var out [][2]common.Address
	for ad, toks := range m {
		a := common.HexToAddress(ad)
		for _, t := range toks {
			out = append(out, [2]common.Address{common.HexToAddress(t), a})
		}
	}
	return out
}

func explodePairs(p2bal map[[2]common.Address]*big.Int) map[string][]dep.TokenBalance {
	out := make(map[string][]dep.TokenBalance)
	for pair, bal := range p2bal {
		token := pair[0].Hex()
		addr := pair[1].Hex()
		out[addr] = append(out[addr], dep.TokenBalance{
			Contract: token,
			NativeBalance: dep.NativeBalance{
				Symbol:   "",
				Decimals: 18,
				Amount:   new(big.Int).Set(bal),
			},
		})
	}
	return out
}

func hashStrings(ss []string) string {
	h := sha3.NewLegacyKeccak256()
	for _, s := range ss {
		h.Write([]byte(s))
	}
	return hex.EncodeToString(h.Sum(nil))[:16]
}

func hashPairs(p [][2]common.Address) string {
	h := sha3.NewLegacyKeccak256()
	for _, x := range p {
		h.Write(x[0].Bytes())
		h.Write(x[1].Bytes())
	}
	return hex.EncodeToString(h.Sum(nil))[:16]
}

func erc20BalanceOfSelector() []byte {
	h := sha3.NewLegacyKeccak256()
	h.Write([]byte("balanceOf(address)"))
	return h.Sum(nil)[:4]
}

// ---------------- Multicall 地址 ----------------

func resolveMulticall(network string) string {
	switch strings.ToUpper(network) {
	case "ETH", "ETH_MAINNET", "MAINNET":
		return "0x5ba1e12693dc8f9c48aad8770482f4739beed696" // Multicall2 on ETH
	case "BSC":
		return "0xca11bde05977b3631167028862be2a173976ca11"
	case "BSC_TESTNET":
		return "0xca11bde05977b3631167028862be2a173976ca11"
	// 如需：POLYGON/ARBITRUM/OP/Base 等在此补充
	default:
		return "" // 为空则走降级路径
	}
}

func (c *EVMClient) multicallBalanceOf(ctx context.Context, network string, pairs [][2]common.Address, a dep.AnchorRef) (map[[2]common.Address]*big.Int, error) {
	rc, _, err := c.pick(network)
	if err != nil {
		return nil, err
	}
	target := common.HexToAddress(c.getMulticallAddr(network))

	type Call struct {
		Target   common.Address
		CallData []byte
	}
	selector := erc20BalanceOfSelector()
	encode := func(p [2]common.Address) []byte {
		data := make([]byte, 4+32)
		copy(data[:4], selector)
		copy(data[4+12:], p[1].Bytes()) // owner
		return data
	}

	out := make(map[[2]common.Address]*big.Int, len(pairs))
	for _, batch := range chunkPairs(pairs, c.maxBatch) {
		calls := make([]Call, 0, len(batch))
		for _, p := range batch {
			calls = append(calls, Call{Target: p[0], CallData: encode(p)})
		}
		input, err := c.multicallABI.Pack("tryAggregate", false, calls)
		if err != nil {
			return nil, err
		}
		msg := map[string]any{"to": target, "data": "0x" + hex.EncodeToString(input)}
		var result string
		ctx2, cancel := context.WithTimeout(ctx, c.reqTimeout)
		err = rc.CallContext(ctx2, &result, "eth_call", msg, blockParam(a))
		cancel()
		if err != nil {
			return nil, err
		}
		raw := common.FromHex(result)
		var decoded []struct {
			Success bool
			Ret     []byte
		}
		if err := c.multicallABI.UnpackIntoInterface(&decoded, "tryAggregate", raw); err != nil {
			return nil, err
		}
		if len(decoded) != len(batch) {
			return nil, fmt.Errorf("multicall size mismatch")
		}
		for i, p := range batch {
			if !decoded[i].Success || len(decoded[i].Ret) < 32 {
				out[p] = big.NewInt(0)
				continue
			}
			out[p] = new(big.Int).SetBytes(decoded[i].Ret[len(decoded[i].Ret)-32:])
		}
	}
	return out, nil
}

func (c *EVMClient) batchedBalanceOf(ctx context.Context, network string, pairs [][2]common.Address, a dep.AnchorRef) (map[[2]common.Address]*big.Int, error) {
	rc, _, err := c.pick(network)
	if err != nil {
		return nil, err
	}
	out := make(map[[2]common.Address]*big.Int, len(pairs))
	selector := erc20BalanceOfSelector()

	for _, batch := range chunkPairs(pairs, c.maxBatch) {
		elems := make([]gethrpc.BatchElem, 0, len(batch))
		results := make([]string, len(batch))
		for i, p := range batch {
			data := make([]byte, 4+32)
			copy(data[:4], selector)
			copy(data[4+12:], p[1].Bytes()) // owner
			msg := map[string]any{"to": p[0], "data": "0x" + hex.EncodeToString(data)}
			elems = append(elems, gethrpc.BatchElem{
				Method: "eth_call",
				Args:   []any{msg, blockParam(a)},
				Result: &results[i],
			})
		}
		ctx2, cancel := context.WithTimeout(ctx, c.reqTimeout)
		err := rc.BatchCallContext(ctx2, elems)
		cancel()
		if err != nil {
			return nil, err
		}
		for i, p := range batch {
			raw := common.FromHex(results[i])
			if len(raw) < 32 {
				out[p] = big.NewInt(0)
				continue
			}
			out[p] = new(big.Int).SetBytes(raw[len(raw)-32:])
		}
	}
	return out, nil
}
