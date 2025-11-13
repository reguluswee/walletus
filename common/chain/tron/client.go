package tron

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/mr-tron/base58"
	"github.com/reguluswee/walletus/common/chain/dep"
	"github.com/reguluswee/walletus/common/config"
	"golang.org/x/crypto/sha3"
	"golang.org/x/sync/singleflight"
)

type httpClient struct {
	baseURL string
	client  *http.Client
	name    string
}

type TRXClient struct {
	mu         sync.RWMutex
	clients    map[string]*httpClient // network -> client
	sf         singleflight.Group
	maxBatch   int
	reqTimeout time.Duration
}

// NewTRXClient 创建一个新的 TRON 客户端实例
// RPC 配置会在首次使用时从系统配置中加载（通过 ensureClient 方法）
func NewTRXClient() *TRXClient {
	return &TRXClient{
		clients:    make(map[string]*httpClient),
		maxBatch:   256,             // 默认批处理大小
		reqTimeout: 5 * time.Second, // TRON RPC 可能需要更长的超时时间
	}
}

func (c *TRXClient) pick(network string) (*httpClient, error) {
	client, err := c.ensureClient(network)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// ensureClient 确保指定网络的 HTTP 客户端已初始化
// 该方法会从系统配置中读取 RPC 端点配置（通过 config.GetRpcConfig）
func (c *TRXClient) ensureClient(network string) (*httpClient, error) {
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
	var httpCli *httpClient
	for _, url := range cc.GetRpc() {
		baseURL := strings.TrimSuffix(url, "/")
		httpCli = &httpClient{
			baseURL: baseURL,
			client: &http.Client{
				Timeout: c.reqTimeout,
			},
			name: shortAlias(network, url),
		}
		// 测试连接（可选，这里简化处理）
		break
	}

	if httpCli == nil {
		return nil, fmt.Errorf("no RPC available for %s", network)
	}

	c.clients[network] = httpCli
	return httpCli, nil
}

// callRPC 调用 TRON RPC API
// TRON API 使用直接 POST JSON 到 /wallet/{method}，请求体是参数对象
func (cli *httpClient) callRPC(ctx context.Context, method string, params interface{}) (map[string]interface{}, error) {
	var bodyBytes []byte
	var err error

	if params == nil {
		bodyBytes = []byte("{}")
	} else {
		bodyBytes, err = json.Marshal(params)
		if err != nil {
			return nil, fmt.Errorf("marshal request: %w", err)
		}
	}

	url := cli.baseURL + "/" + method
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bodyBytes))
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

	// TRON API 可能返回错误字符串或 JSON 对象
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http status %d: %s", resp.StatusCode, string(respBody))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		// 如果不是 JSON，可能是错误字符串
		return nil, fmt.Errorf("unmarshal response: %w, body: %s", err, string(respBody))
	}

	// 检查错误字段
	if errMsg, ok := result["Error"]; ok && errMsg != nil {
		return nil, fmt.Errorf("rpc error: %v", errMsg)
	}

	return result, nil
}

func (c *TRXClient) Anchor(ctx context.Context, network string, cs dep.Consistency) (dep.AnchorRef, error) {
	cli, err := c.pick(network)
	if err != nil {
		return dep.AnchorRef{}, err
	}

	ctx2, cancel := context.WithTimeout(ctx, c.reqTimeout)
	defer cancel()

	tag := cs.Mode
	if tag == "" {
		tag = "latest"
	}

	var blockNum int64
	switch tag {
	case "latest":
		// 获取最新区块
		resp, err := cli.callRPC(ctx2, "wallet/getnowblock", nil)
		if err != nil {
			return dep.AnchorRef{}, err
		}
		if blockHeader, ok := resp["block_header"].(map[string]interface{}); ok {
			if rawData, ok := blockHeader["raw_data"].(map[string]interface{}); ok {
				if num, ok := rawData["number"].(float64); ok {
					blockNum = int64(num)
				}
			}
		}
	case "latest_solid":
		// 获取最新确认区块（可能需要多次调用或使用特定 API）
		resp, err := cli.callRPC(ctx2, "wallet/getnowblock", nil)
		if err != nil {
			return dep.AnchorRef{}, err
		}
		if blockHeader, ok := resp["block_header"].(map[string]interface{}); ok {
			if rawData, ok := blockHeader["raw_data"].(map[string]interface{}); ok {
				if num, ok := rawData["number"].(float64); ok {
					blockNum = int64(num)
				}
			}
		}
	default:
		// 尝试解析为区块号
		if num, err := parseBlockNumber(tag); err == nil {
			blockNum = num
		} else {
			return dep.AnchorRef{}, fmt.Errorf("invalid block tag: %s", tag)
		}
	}

	return dep.AnchorRef{
		Height:   uint64(blockNum),
		Tag:      cs.Mode,
		Network:  network,
		Provider: cli.name,
	}, nil
}

func (c *TRXClient) NativeBalance(ctx context.Context, network, address string, a dep.AnchorRef) (*dep.NativeBalance, error) {
	mp, err := c.NativeBalanceBatch(ctx, network, []string{address}, a)
	if err != nil {
		return nil, err
	}
	return mp[address], nil
}

func (c *TRXClient) TokenBalances(ctx context.Context, network, address string, tokens []string, a dep.AnchorRef) ([]dep.TokenBalance, error) {
	mp, err := c.TokenBalancesBatch(ctx, network, map[string][]string{address: tokens}, a)
	if err != nil {
		return nil, err
	}
	return mp[address], nil
}

func (c *TRXClient) NativeBalanceBatch(ctx context.Context, network string, addrs []string, a dep.AnchorRef) (map[string]*dep.NativeBalance, error) {
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
		for _, batch := range chunkStrings(addrs, c.maxBatch) {
			for _, addr := range batch {
				balance, err := c.getAccountBalance(ctx, cli, addr)
				if err != nil {
					// 如果账户不存在或出错，返回零余额
					out[addr] = &dep.NativeBalance{
						Symbol:   "TRX",
						Decimals: 6,
						Amount:   big.NewInt(0),
					}
					continue
				}
				out[addr] = balance
			}
		}
		return out, nil
	})

	if err != nil {
		return nil, err
	}
	return v.(map[string]*dep.NativeBalance), nil
}

func (c *TRXClient) TokenBalancesBatch(ctx context.Context, network string, addr2tokens map[string][]string, a dep.AnchorRef) (map[string][]dep.TokenBalance, error) {
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
		for addr, tokens := range addr2tokens {
			if len(tokens) == 0 {
				out[addr] = []dep.TokenBalance{}
				continue
			}

			var tokenBalances []dep.TokenBalance
			for _, tokenAddr := range tokens {
				balance, err := c.getTRC20Balance(ctx, cli, tokenAddr, addr)
				if err != nil {
					// 如果查询失败，返回零余额
					tokenBalances = append(tokenBalances, dep.TokenBalance{
						Contract: tokenAddr,
						NativeBalance: dep.NativeBalance{
							Symbol:   "",
							Decimals: 18, // 默认值，实际可能需要查询合约
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

// getAccountBalance 获取账户的 TRX 余额
func (c *TRXClient) getAccountBalance(ctx context.Context, cli *httpClient, address string) (*dep.NativeBalance, error) {
	// TRON API 支持 visible=true，可以直接使用 base58 地址
	params := map[string]interface{}{
		"address": address,
		"visible": true,
	}

	resp, err := cli.callRPC(ctx, "wallet/getaccount", params)
	if err != nil {
		return nil, err
	}

	var balance int64
	if balanceVal, ok := resp["balance"]; ok {
		switch v := balanceVal.(type) {
		case float64:
			balance = int64(v)
		case int64:
			balance = v
		case string:
			// 尝试解析字符串
			if b, err := parseStringToInt64(v); err == nil {
				balance = b
			}
		}
	}

	return &dep.NativeBalance{
		Symbol:   "TRX",
		Decimals: 6,
		Amount:   big.NewInt(balance),
	}, nil
}

// getTRC20Balance 获取 TRC20 Token 余额
func (c *TRXClient) getTRC20Balance(ctx context.Context, cli *httpClient, tokenAddr, ownerAddr string) (*dep.TokenBalance, error) {
	// 编码 balanceOf(address) 的参数
	// 函数选择器：balanceOf(address) 的 keccak256 前4字节
	// 参数：address 填充到 32 字节（左侧补0）
	parameter, err := encodeTRC20BalanceOfParameter(ownerAddr)
	if err != nil {
		return nil, fmt.Errorf("encode parameter: %w", err)
	}

	params := map[string]interface{}{
		"owner_address":     ownerAddr,
		"contract_address":  tokenAddr,
		"function_selector": "balanceOf(address)",
		"parameter":         parameter,
		"visible":           true,
	}

	resp, err := cli.callRPC(ctx, "wallet/triggerconstantcontract", params)
	if err != nil {
		return nil, err
	}

	var balance *big.Int = big.NewInt(0)
	if constantResult, ok := resp["constant_result"].([]interface{}); ok && len(constantResult) > 0 {
		if resultHex, ok := constantResult[0].(string); ok {
			// 解码 hex 字符串为 big.Int
			resultBytes, err := hex.DecodeString(resultHex)
			if err == nil && len(resultBytes) >= 32 {
				balance = new(big.Int).SetBytes(resultBytes[len(resultBytes)-32:])
			}
		}
	}

	return &dep.TokenBalance{
		Contract: tokenAddr,
		NativeBalance: dep.NativeBalance{
			Symbol:   "", // 需要额外查询获取 symbol
			Decimals: 18, // 需要额外查询获取 decimals，默认值
			Amount:   balance,
		},
	}, nil
}

// encodeTRC20BalanceOfParameter 编码 balanceOf(address) 的参数
func encodeTRC20BalanceOfParameter(address string) (string, error) {
	// 将 base58 地址转换为 hex 格式（20 字节）
	addrBytes, err := base58AddressToHex(address)
	if err != nil {
		return "", fmt.Errorf("convert address: %w", err)
	}

	// 填充到 32 字节（左侧补0）
	param := make([]byte, 32)
	copy(param[32-len(addrBytes):], addrBytes)

	return hex.EncodeToString(param), nil
}

// base58AddressToHex 将 TRON base58 地址转换为 hex 格式（去掉 0x41 前缀和校验和）
func base58AddressToHex(base58Addr string) ([]byte, error) {
	// 解码 base58
	decoded, err := base58.Decode(base58Addr)
	if err != nil {
		return nil, fmt.Errorf("decode base58: %w", err)
	}

	// TRON 地址格式：0x41 (1 byte) + address (20 bytes) + checksum (4 bytes)
	if len(decoded) != 25 {
		return nil, fmt.Errorf("invalid address length: %d", len(decoded))
	}

	if decoded[0] != 0x41 {
		return nil, fmt.Errorf("invalid address prefix: 0x%02x", decoded[0])
	}

	// 提取中间的 20 字节地址
	addressBytes := decoded[1:21]

	// 验证校验和（可选，但推荐）
	checksum := decoded[21:]
	raw := decoded[:21]
	sum := sha256.Sum256(raw)
	sum = sha256.Sum256(sum[:])
	expectedChecksum := sum[:4]
	for i := 0; i < 4; i++ {
		if checksum[i] != expectedChecksum[i] {
			return nil, fmt.Errorf("invalid checksum")
		}
	}

	return addressBytes, nil
}

func MustRegister() {
	dep.Register(dep.GetSupportedTron(), NewTRXClient())
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
	var parts []string
	for addr, tokens := range m {
		parts = append(parts, fmt.Sprintf("%s:%s", addr, strings.Join(tokens, ",")))
	}
	h := strings.Join(parts, "|")
	return fmt.Sprintf("%x", h)[:16]
}

// parseBlockNumber 解析区块号
func parseBlockNumber(s string) (int64, error) {
	// 尝试解析为十进制数字
	num := new(big.Int)
	if _, ok := num.SetString(s, 10); !ok {
		return 0, fmt.Errorf("invalid block number format: %s", s)
	}
	if !num.IsInt64() {
		return 0, fmt.Errorf("block number too large: %s", s)
	}
	return num.Int64(), nil
}

// parseStringToInt64 解析字符串为 int64
func parseStringToInt64(s string) (int64, error) {
	num := new(big.Int)
	if _, ok := num.SetString(s, 10); !ok {
		return 0, fmt.Errorf("invalid number format: %s", s)
	}
	if !num.IsInt64() {
		return 0, fmt.Errorf("number too large: %s", s)
	}
	return num.Int64(), nil
}
