package request

type WalletCreateRequest struct {
	TenantID uint64 `json:"tenant_id"`
	UniqueID uint32 `json:"unique_id"`
	Chain    string `json:"chain"`
}

type TenantCreateRequest struct {
	Name     string `json:"name"`
	UniqueID string `json:"unique_id"`
	Callback string `json:"call_back"`
}

type WalletBalanceQueryRequest struct {
	TenantID  uint64 `json:"tenant_id"`
	AddressID uint64 `json:"address_id"`
	Address   string `json:"address"`
	Chain     string `json:"chain"`
	Token     string `json:"token"`
}
