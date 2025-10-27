package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestConfigInit(t *testing.T) {
	rpcList := systemConfig.Chain[0].GetRpc()
	rpcMapList := systemConfig.Chain[0].GetRpcMapper()
	fmt.Println(rpcList, rpcMapList)

	i := systemConfig.Chain[0].RpcMap[rpcList[0]]
	fmt.Println(i)
}

func TestJitoRPC(t *testing.T) {
	var url = "https://mainnet.block-engine.jito.wtf/api/v1/transactions"
	t.Logf("jito rpc %s", url)

	// 构造请求数据
	payload := map[string]interface{}{
		"id":      1,
		"jsonrpc": "2.0",
		"method":  "sendTransaction",
		"params": []interface{}{
			"AVXo5X7UNzpuOmYzkZ+fqHDGiRLTSMlWlUCcZKzEV5CIKlrdvZa3/2GrJJfPrXgZqJbYDaGiOnP99tI/sRJfiwwBAAEDRQ/n5E5CLbMbHanUG3+iVvBAWZu0WFM6NoB5xfybQ7kNwwgfIhv6odn2qTUu/gOisDtaeCW1qlwW/gx3ccr/4wAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAvsInicc+E3IZzLqeA+iM5cn9kSaeFzOuClz1Z2kZQy0BAgIAAQwCAAAAAPIFKgEAAAA=",
			map[string]string{"encoding": "base64"},
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	fmt.Println("Response:", result)
}
