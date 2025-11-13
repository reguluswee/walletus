package chain

import (
	"context"
	"fmt"
	"math/big"
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
		Chain:     eth,
		Network:   "mainnet",
		Addresses: []string{"0xe38533e11B680eAf4C9519Ea99B633BD3ef5c2F8"},
		Tokens: map[string][]string{
			"0xe38533e11B680eAf4C9519Ea99B633BD3ef5c2F8": {"0xD9A442856C234a39a81a089C06451EBAa4306a72"},
		},
		Consistency: dep.Consistency{Mode: "safe", MinConfirmations: 0},
	}
	res, err := gw.GetBalances(context.Background(), q)
	if err != nil {
		t.Fatalf("查询余额失败: %v", err)
	}

	printBatchBalanceResult(t, res)

	fmt.Println("\n========== 余额查询结果（控制台输出）==========")
	printBatchBalanceResultToConsole(res)
}

// printBatchBalanceResult 格式化打印 BatchBalanceResult
func printBatchBalanceResult(t *testing.T, res *dep.BatchBalanceResult) {
	t.Log("========== 余额查询结果 ==========")
	t.Logf("链名称: %s (CoinType: %d)", res.Chain.Name, res.Chain.CoinType)
	t.Logf("查询结果数量: %d", len(res.Results))
	t.Log("")

	for i, result := range res.Results {
		t.Logf("------ 地址 %d: %s ------", i+1, result.Address)
		t.Logf("查询时间: %s", result.QueriedAtUTC)
		t.Logf("区块高度: %d", result.Anchor.Height)
		t.Logf("区块标签: %s", result.Anchor.Tag)
		t.Logf("网络: %s", result.Anchor.Network)
		t.Logf("RPC 提供商: %s", result.Anchor.Provider)
		t.Log("")

		// 原生币余额
		if result.Native != nil {
			amount := formatBalance(result.Native.Amount, result.Native.Decimals)
			t.Logf("原生币 (%s):", result.Native.Symbol)
			t.Logf("  余额: %s %s", amount, result.Native.Symbol)
			t.Logf("  原始值 (Wei): %s", result.Native.Amount.String())
			t.Logf("  小数位数: %d", result.Native.Decimals)
		} else {
			t.Log("原生币: 无余额")
		}
		t.Log("")

		// Token 余额
		if len(result.Tokens) > 0 {
			t.Logf("Token 数量: %d", len(result.Tokens))
			for j, token := range result.Tokens {
				amount := formatBalance(token.Amount, token.Decimals)
				t.Logf("  Token %d:", j+1)
				t.Logf("    合约地址: %s", token.Contract)
				t.Logf("    符号: %s", token.Symbol)
				t.Logf("    余额: %s %s", amount, token.Symbol)
				t.Logf("    原始值: %s", token.Amount.String())
				t.Logf("    小数位数: %d", token.Decimals)
			}
		} else {
			t.Log("Token: 无余额")
		}
		t.Log("")
	}
	t.Log("==================================")
}

// formatBalance 将 big.Int 余额格式化为可读的字符串（考虑小数位数）
func formatBalance(amount *big.Int, decimals int) string {
	if amount == nil {
		return "0"
	}

	// 创建 10^decimals 的除数
	divisor := new(big.Int)
	divisor.Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)

	// 计算整数部分和小数部分
	quotient := new(big.Int)
	remainder := new(big.Int)
	quotient.DivMod(amount, divisor, remainder)

	// 格式化为字符串
	if remainder.Sign() == 0 {
		// 没有小数部分
		return quotient.String()
	}

	// 格式化小数部分（去掉尾部的 0）
	remainderStr := remainder.String()
	// 补齐前导零
	for len(remainderStr) < decimals {
		remainderStr = "0" + remainderStr
	}
	// 去掉尾部零
	remainderStr = trimTrailingZeros(remainderStr)

	return fmt.Sprintf("%s.%s", quotient.String(), remainderStr)
}

// trimTrailingZeros 去掉字符串尾部的零
func trimTrailingZeros(s string) string {
	for len(s) > 0 && s[len(s)-1] == '0' {
		s = s[:len(s)-1]
	}
	return s
}

// printBatchBalanceResultToConsole 格式化打印到控制台（使用 fmt.Printf）
func printBatchBalanceResultToConsole(res *dep.BatchBalanceResult) {
	fmt.Printf("链名称: %s (CoinType: %d)\n", res.Chain.Name, res.Chain.CoinType)
	fmt.Printf("查询结果数量: %d\n\n", len(res.Results))

	for i, result := range res.Results {
		fmt.Printf("------ 地址 %d: %s ------\n", i+1, result.Address)
		fmt.Printf("查询时间: %s\n", result.QueriedAtUTC)
		fmt.Printf("区块高度: %d\n", result.Anchor.Height)
		fmt.Printf("区块标签: %s\n", result.Anchor.Tag)
		fmt.Printf("网络: %s\n", result.Anchor.Network)
		fmt.Printf("RPC 提供商: %s\n\n", result.Anchor.Provider)

		// 原生币余额
		if result.Native != nil {
			amount := formatBalance(result.Native.Amount, result.Native.Decimals)
			fmt.Printf("原生币 (%s):\n", result.Native.Symbol)
			fmt.Printf("  余额: %s %s\n", amount, result.Native.Symbol)
			fmt.Printf("  原始值 (Wei): %s\n", result.Native.Amount.String())
			fmt.Printf("  小数位数: %d\n\n", result.Native.Decimals)
		} else {
			fmt.Printf("原生币: 无余额\n\n")
		}

		// Token 余额
		if len(result.Tokens) > 0 {
			fmt.Printf("Token 数量: %d\n", len(result.Tokens))
			for j, token := range result.Tokens {
				amount := formatBalance(token.Amount, token.Decimals)
				fmt.Printf("  Token %d:\n", j+1)
				fmt.Printf("    合约地址: %s\n", token.Contract)
				fmt.Printf("    符号: %s\n", token.Symbol)
				fmt.Printf("    余额: %s %s\n", amount, token.Symbol)
				fmt.Printf("    原始值: %s\n", token.Amount.String())
				fmt.Printf("    小数位数: %d\n", token.Decimals)
			}
		} else {
			fmt.Printf("Token: 无余额\n")
		}
		fmt.Println()
	}
	fmt.Println("==================================")
}
