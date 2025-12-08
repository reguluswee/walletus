package solana

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/reguluswee/walletus/common/chain/dep"
	"github.com/reguluswee/walletus/common/config"
	"golang.org/x/crypto/sha3"
	"golang.org/x/sync/singleflight"
)

type rpcClient struct {
	baseURL string
	client  *http.Client
	name    string
}

type SOLClient struct {
	mu         sync.RWMutex
	clients    map[string]*rpcClient // network -> client
	sf         singleflight.Group
	maxBatch   int
	reqTimeout time.Duration
}

// NewSOLClient 创建一个新的 Solana 客户端实例
// RPC 配置会在首次使用时从系统配置中加载（通过 ensureClient 方法）
func NewSOLClient() *SOLClient {
	return &SOLClient{
		clients:    make(map[string]*rpcClient),
		maxBatch:   256,             // 默认批处理大小
		reqTimeout: 5 * time.Second, // Solana RPC 可能需要更长的超时时间
	}
}

func (c *SOLClient) pick(network string) (*rpcClient, error) {
	client, err := c.ensureClient(network)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// ensureClient 确保指定网络的 HTTP 客户端已初始化
// 该方法会从系统配置中读取 RPC 端点配置（通过 config.GetRpcConfig）
func (c *SOLClient) ensureClient(network string) (*rpcClient, error) {
	c.mu.RLock()
	client, ok := c.clients[network]
	c.mu.RUnlock()
	if ok {
		return client, nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	// 双检锁，避免并发创建
	if client, ok := c.clients[network]; ok {
		return client, nil
	}

	// 从系统配置中获取 RPC 配置
	cc := config.GetRpcConfig(network)
	if cc == nil || len(cc.GetRpc()) == 0 {
		return nil, fmt.Errorf("missing RPC config for %s", network)
	}

	// 取第一个可用的 RPC 端点
	var rpcCli *rpcClient
	for _, url := range cc.GetRpc() {
		baseURL := strings.TrimSuffix(url, "/")
		rpcCli = &rpcClient{
			baseURL: baseURL,
			client: &http.Client{
				Timeout: c.reqTimeout,
			},
			name: shortAlias(network, url),
		}
		break
	}

	if rpcCli == nil {
		return nil, fmt.Errorf("no RPC available for %s", network)
	}

	c.clients[network] = rpcCli
	return rpcCli, nil
}

// callRPC 调用 Solana JSON-RPC API
func (cli *rpcClient) callRPC(ctx context.Context, method string, params []interface{}) (json.RawMessage, error) {
	reqBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  method,
		"params":  params,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", cli.baseURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := cli.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http status %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		JSONRPC string          `json:"jsonrpc"`
		ID      int             `json:"id"`
		Result  json.RawMessage `json:"result"`
		Error   *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w, body: %s", err, string(respBody))
	}

	if result.Error != nil {
		return nil, fmt.Errorf("rpc error %d: %s", result.Error.Code, result.Error.Message)
	}

	return result.Result, nil
}

func (c *SOLClient) Anchor(ctx context.Context, network string, cs dep.Consistency) (dep.AnchorRef, error) {
	cli, err := c.pick(network)
	if err != nil {
		return dep.AnchorRef{}, err
	}

	ctx2, cancel := context.WithTimeout(ctx, c.reqTimeout)
	defer cancel()

	tag := cs.Mode
	if tag == "" {
		tag = "confirmed"
	}

	// 根据一致性模式选择 commitment level
	commitment := tag
	switch tag {
	case "processed", "confirmed", "finalized":
		commitment = tag
	default:
		commitment = "confirmed"
	}

	// 获取 Slot
	params := []interface{}{
		map[string]interface{}{
			"commitment": commitment,
		},
	}

	var slot uint64
	result, err := cli.callRPC(ctx2, "getSlot", params)
	if err != nil {
		return dep.AnchorRef{}, err
	}

	if err := json.Unmarshal(result, &slot); err != nil {
		return dep.AnchorRef{}, fmt.Errorf("unmarshal slot: %w", err)
	}

	return dep.AnchorRef{
		Height:   slot,
		Tag:      cs.Mode,
		Network:  network,
		Provider: cli.name,
	}, nil
}

func (c *SOLClient) NativeBalance(ctx context.Context, network, address string, a dep.AnchorRef) (*dep.NativeBalance, error) {
	mp, err := c.NativeBalanceBatch(ctx, network, []string{address}, a)
	if err != nil {
		return nil, err
	}
	return mp[address], nil
}

func (c *SOLClient) TokenBalances(ctx context.Context, network, address string, tokens []string, a dep.AnchorRef) ([]dep.TokenBalance, error) {
	mp, err := c.TokenBalancesBatch(ctx, network, map[string][]string{address: tokens}, a)
	if err != nil {
		return nil, err
	}
	return mp[address], nil
}

func (c *SOLClient) NativeBalanceBatch(ctx context.Context, network string, addrs []string, a dep.AnchorRef) (map[string]*dep.NativeBalance, error) {
	if len(addrs) == 0 {
		return map[string]*dep.NativeBalance{}, nil
	}

	key := fmt.Sprintf("nb:%s:%d:%s", network, a.Height, hashStrings(addrs))
	v, err, _ := c.sf.Do(key, func() (interface{}, error) {
		cli, err := c.pick(network)
		if err != nil {
			return nil, err
		}

		out := make(map[string]*dep.NativeBalance, len(addrs))

		// Solana 支持批量查询，使用 getMultipleAccounts
		for _, batch := range chunkStrings(addrs, c.maxBatch) {
			params := []interface{}{
				batch,
				map[string]interface{}{
					"commitment": getCommitmentFromTag(a.Tag),
				},
			}

			result, err := cli.callRPC(ctx, "getMultipleAccounts", params)
			if err != nil {
				// 如果批量查询失败，降级为单个查询
				for _, addr := range batch {
					balance, e := c.getAccountBalance(ctx, cli, addr, a.Tag)
					if e != nil {
						out[addr] = &dep.NativeBalance{
							Symbol:   "SOL",
							Decimals: 9,
							Amount:   big.NewInt(0),
						}
						continue
					}
					out[addr] = balance
				}
				continue
			}

			var accounts struct {
				Value []*struct {
					Data       interface{} `json:"data"`
					Executable bool        `json:"executable"`
					Lamports   uint64      `json:"lamports"`
					Owner      string      `json:"owner"`
					RentEpoch  uint64      `json:"rentEpoch"`
				} `json:"value"`
			}

			if err := json.Unmarshal(result, &accounts); err != nil {
				// 解析失败，降级为单个查询
				for _, addr := range batch {
					balance, e := c.getAccountBalance(ctx, cli, addr, a.Tag)
					if e != nil {
						out[addr] = &dep.NativeBalance{
							Symbol:   "SOL",
							Decimals: 9,
							Amount:   big.NewInt(0),
						}
						continue
					}
					out[addr] = balance
				}
				continue
			}

			for i, addr := range batch {
				if i < len(accounts.Value) && accounts.Value[i] != nil && accounts.Value[i].Lamports > 0 {
					out[addr] = &dep.NativeBalance{
						Symbol:   "SOL",
						Decimals: 9,
						Amount:   big.NewInt(int64(accounts.Value[i].Lamports)),
					}
				} else {
					// 账户不存在或余额为0
					out[addr] = &dep.NativeBalance{
						Symbol:   "SOL",
						Decimals: 9,
						Amount:   big.NewInt(0),
					}
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

func (c *SOLClient) TokenBalancesBatch(ctx context.Context, network string, addr2tokens map[string][]string, a dep.AnchorRef) (map[string][]dep.TokenBalance, error) {
	if len(addr2tokens) == 0 {
		return map[string][]dep.TokenBalance{}, nil
	}

	key := fmt.Sprintf("tb:%s:%d:%s", network, a.Height, hashAddr2Tokens(addr2tokens))
	v, err, _ := c.sf.Do(key, func() (interface{}, error) {
		cli, err := c.pick(network)
		if err != nil {
			return nil, err
		}

		out := make(map[string][]dep.TokenBalance)
		for addr, tokenMints := range addr2tokens {
			if len(tokenMints) == 0 {
				out[addr] = []dep.TokenBalance{}
				continue
			}

			var tokenBalances []dep.TokenBalance
			for _, tokenMint := range tokenMints {
				balance, err := c.getSPLTokenBalance(ctx, cli, addr, tokenMint, a.Tag)
				if err != nil {
					// 如果查询失败，返回零余额
					tokenBalances = append(tokenBalances, dep.TokenBalance{
						Contract: tokenMint,
						NativeBalance: dep.NativeBalance{
							Symbol:   "",
							Decimals: 0, // 默认值，实际需要查询 Token 元数据
							Amount:   big.NewInt(0),
						},
					})
					continue
				}
				tokenBalances = append(tokenBalances, *balance)
			}
			out[addr] = tokenBalances
		}
		return out, nil
	})

	if err != nil {
		return nil, err
	}
	return v.(map[string][]dep.TokenBalance), nil
}

// getAccountBalance 获取账户的 SOL 余额
func (c *SOLClient) getAccountBalance(ctx context.Context, cli *rpcClient, address, commitment string) (*dep.NativeBalance, error) {
	params := []interface{}{
		address,
		map[string]interface{}{
			"commitment": getCommitmentFromTag(commitment),
		},
	}

	result, err := cli.callRPC(ctx, "getBalance", params)
	if err != nil {
		return nil, err
	}

	var balanceResp struct {
		Value uint64 `json:"value"`
	}

	if err := json.Unmarshal(result, &balanceResp); err != nil {
		return nil, fmt.Errorf("unmarshal balance: %w", err)
	}

	return &dep.NativeBalance{
		Symbol:   "SOL",
		Decimals: 9,
		Amount:   big.NewInt(int64(balanceResp.Value)),
	}, nil
}

// getSPLTokenBalance 获取 SPL Token 余额
func (c *SOLClient) getSPLTokenBalance(ctx context.Context, cli *rpcClient, ownerAddr, tokenMint, commitment string) (*dep.TokenBalance, error) {
	// 使用 getTokenAccountsByOwner 查询 Token 账户
	params := []interface{}{
		ownerAddr,
		map[string]interface{}{
			"mint": tokenMint,
		},
		map[string]interface{}{
			"encoding":   "jsonParsed",
			"commitment": getCommitmentFromTag(commitment),
		},
	}

	result, err := cli.callRPC(ctx, "getTokenAccountsByOwner", params)
	if err != nil {
		return nil, err
	}

	var accountsResp struct {
		Value []struct {
			Account struct {
				Data struct {
					Parsed struct {
						Info struct {
							TokenAmount struct {
								Amount         string `json:"amount"`
								Decimals       uint8  `json:"decimals"`
								UIAmountString string `json:"uiAmountString"`
							} `json:"tokenAmount"`
							Mint string `json:"mint"`
						} `json:"info"`
					} `json:"parsed"`
				} `json:"data"`
			} `json:"account"`
		} `json:"value"`
	}

	if err := json.Unmarshal(result, &accountsResp); err != nil {
		return nil, fmt.Errorf("unmarshal token accounts: %w", err)
	}

	var totalAmount *big.Int = big.NewInt(0)
	var decimals uint8 = 0

	for _, account := range accountsResp.Value {
		tokenAmount := account.Account.Data.Parsed.Info.TokenAmount
		if tokenAmount.Amount != "" {
			amount := new(big.Int)
			if _, ok := amount.SetString(tokenAmount.Amount, 10); ok {
				totalAmount.Add(totalAmount, amount)
			}
		}
		if decimals == 0 && tokenAmount.Decimals > 0 {
			decimals = tokenAmount.Decimals
		}
	}

	return &dep.TokenBalance{
		Contract: tokenMint,
		NativeBalance: dep.NativeBalance{
			Symbol:   "", // 需要额外查询 Token 元数据获取 symbol
			Decimals: int(decimals),
			Amount:   totalAmount,
		},
	}, nil
}

func (c *SOLClient) GetTransaction(ctx context.Context, network string, txHash string) (any, error) {
	return nil, errors.New("not implemented")
}

func MustRegister() {
	dep.Register(dep.GetSupportedSol(), NewSOLClient())
}

// ---------------- 工具函数 ----------------

// shortAlias 生成 RPC 提供商的简短别名
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

// chunkStrings 将字符串切片分块
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

// hashStrings 计算字符串切片的哈希值
func hashStrings(ss []string) string {
	h := sha3.NewLegacyKeccak256()
	for _, s := range ss {
		h.Write([]byte(s))
	}
	return hex.EncodeToString(h.Sum(nil))[:16]
}

// hashAddr2Tokens 计算地址到 Token 列表映射的哈希值
func hashAddr2Tokens(m map[string][]string) string {
	h := sha3.NewLegacyKeccak256()
	for addr, tokens := range m {
		h.Write([]byte(addr))
		for _, token := range tokens {
			h.Write([]byte(token))
		}
	}
	return hex.EncodeToString(h.Sum(nil))[:16]
}

// getCommitmentFromTag 根据一致性标签获取 Solana commitment level
func getCommitmentFromTag(tag string) string {
	switch tag {
	case "processed", "confirmed", "finalized":
		return tag
	default:
		return "confirmed"
	}
}
