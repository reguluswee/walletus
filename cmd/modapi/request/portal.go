package request

import "github.com/shopspring/decimal"

type PortalLoginRequest struct {
	LoginID  string `json:"login_id"`
	Password string `json:"password"`
}

type PortalDeptCreateRequest struct {
	Name string `json:"name" binding:"required"`
	Desc string `json:"desc" binding:"required"`
}

type PortalDeptUpdateRequest struct {
	ID   uint64 `json:"id" binding:"required"`
	Name string `json:"name" binding:"required"`
	Desc string `json:"desc" binding:"required"`
}

type PortalUserAddRequest struct {
	Name     string   `json:"name" binding:"required"`
	LoginID  string   `json:"login_id"`
	Email    string   `json:"email"`
	Location string   `json:"location"`
	DeptIDs  []uint64 `json:"dept_ids"`
	Password string   `json:"password"`
}

type PortalUserUpdateRequest struct {
	ID       uint64   `json:"id" binding:"required"`
	Name     string   `json:"name" binding:"required"`
	Email    string   `json:"email" binding:"required"`
	Location string   `json:"location"`
	DeptIDs  []uint64 `json:"dept_ids"`
}

type PortalRoleCreateRequest struct {
	ID   uint64 `json:"id"`
	Name string `json:"name" binding:"required"`
	Desc string `json:"desc"`
}

type PortalPayslipItem struct {
	UserID        uint64          `json:"user_id" binding:"required"`
	WalletID      uint64          `json:"wallet_id"`
	WalletAddress string          `json:"wallet_address" binding:"required"`
	WalletType    string          `json:"wallet_type" binding:"required"`
	WalletChain   string          `json:"wallet_chain" binding:"required"`
	Amount        decimal.Decimal `json:"amount" binding:"required"`
}

type PortalPayrollCreateRequest struct {
	ID          uint64          `json:"id"`
	RollMonth   string          `json:"roll_month"`
	TotalAmount decimal.Decimal `json:"total_amount"`
	Status      string          `json:"status"`
	Desc        string
	Items       []PortalPayslipItem `json:"items"`
	Op          string              `json:"op"`
}

type PortalPayrollStaffWalletRequest struct {
	UserID        uint64 `json:"user_id"`
	WalletAddress string `json:"wallet_address"`
	WalletType    string `json:"wallet_type"`
	WalletChain   string `json:"wallet_chain"`
}

type PortalTenantCreateRequest struct {
	ID       uint64 `json:"id"`
	UniqueID string `json:"unique_id"`
	Name     string `json:"name"`
	Desc     string `json:"desc"`
	Callback string `json:"callback"`
}
