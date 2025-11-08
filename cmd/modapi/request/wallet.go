package request

type WalletCreateRequest struct {
	TenantID uint64 `json:"tenant_id"`
	UniqueID uint32 `json:"unique_id"`
	Chain    string `json:"chain"`
}
