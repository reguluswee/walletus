package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type PortalUser struct {
	ID       uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	LoginID  string    `gorm:"column:login_id;type:varchar(100);not null" json:"login_id"`
	Name     string    `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Email    string    `gorm:"column:email;type:varchar(255);not null" json:"email"`
	Password string    `gorm:"column:password;type:varchar(255);not null" json:"password"`
	AddTime  time.Time `gorm:"column:add_time" json:"add_time"`
	Flag     uint8     `gorm:"column:flag" json:"flag"`
	Location string    `gorm:"column:location;type:varchar(255)" json:"location"`
	Type     string    `gorm:"column:type;type:varchar(255);not null" json:"type"`
}

func (PortalUser) TableName() string {
	return "admin_portal_user"
}

type PortalDept struct {
	ID      uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name    string    `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Desc    string    `gorm:"column:desc;type:varchar(255);not null" json:"desc"`
	AddTime time.Time `gorm:"column:add_time" json:"add_time"`
	Flag    uint8     `gorm:"column:flag" json:"flag"`
	Status  string    `gorm:"column:status" json:"status"`
}

func (PortalDept) TableName() string {
	return "admin_portal_dept"
}

type PortalDeptUser struct {
	ID     uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	DeptID uint64 `gorm:"column:dept_id;not null" json:"dept_id"`
	UserID uint64 `gorm:"column:user_id;not null" json:"user_id"`
}

func (PortalDeptUser) TableName() string {
	return "admin_portal_dept_user"
}

type PortalUserWallet struct {
	ID            uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID        uint64    `gorm:"column:user_id;not null" json:"user_id"`
	WalletAddress string    `gorm:"column:wallet_address;not null" json:"wallet_address"`
	WalletType    string    `gorm:"column:wallet_type;not null" json:"wallet_type"`
	WalletChain   string    `gorm:"column:wallet_chain;not null" json:"wallet_chain"`
	AddTime       time.Time `gorm:"column:add_time" json:"add_time"`
	Flag          uint8     `gorm:"column:flag" json:"flag"`
}

func (PortalUserWallet) TableName() string {
	return "admin_portal_user_wallet"
}

type PortalRole struct {
	ID   uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	Name string `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Flag uint8  `gorm:"column:flag" json:"flag"`
}

func (PortalRole) TableName() string {
	return "admin_portal_role"
}

type PortalFunc struct {
	ID       uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	Name     string `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Flag     uint8  `gorm:"column:flag" json:"flag"`
	ResURI   string `gorm:"column:res_uri;type:varchar(255);not null" json:"res_uri"`
	PermCode string `gorm:"column:perm_code;type:varchar(255);not null" json:"perm_code"`
	Type     string `gorm:"column:type;type:varchar(255);not null" json:"type"`
	Group    string `gorm:"column:group;type:varchar(255);not null" json:"group"`
}

func (PortalFunc) TableName() string {
	return "admin_portal_function"
}

type PortalRoleFunc struct {
	ID     uint64 `gorm:"primaryKey;column:id"`
	RoleID uint64 `gorm:"column:role_id;not null"`
	FuncID uint64 `gorm:"column:func_id;not null"`
}

func (PortalRoleFunc) TableName() string {
	return "admin_portal_role_func"
}

type PortalUserRole struct {
	ID     uint64 `gorm:"primaryKey;column:id"`
	RoleID uint64 `gorm:"column:role_id;not null"`
	UserID uint64 `gorm:"column:user_id;not null"`
}

func (PortalUserRole) TableName() string {
	return "admin_portal_user_role"
}

type PortalPayroll struct {
	ID          uint64          `gorm:"primaryKey;autoIncrement" json:"id"`
	RollMonth   string          `gorm:"column:roll_month;type:varchar(255);not null" json:"roll_month"`
	Flag        uint8           `gorm:"column:flag" json:"flag"`
	CreatorID   uint64          `gorm:"column:creator_id;not null" json:"creator_id"`
	TotalAmount decimal.Decimal `gorm:"column:total_amount;type:decimal(18,2);not null" json:"total_amount"`
	Status      string          `gorm:"column:status;type:varchar(255);not null" json:"status"`
	AddTime     time.Time       `gorm:"column:add_time" json:"add_time"`
	Desc        string          `gorm:"column:desc;type:varchar(255);not null" json:"desc"`
	PayTime     time.Time       `gorm:"column:pay_time" json:"pay_time"`
}

func (PortalPayroll) TableName() string {
	return "admin_portal_payroll"
}

type PortalPayslip struct {
	ID            uint64          `gorm:"primaryKey;autoIncrement" json:"id"`
	PayrollID     uint64          `gorm:"column:payroll_id;not null" json:"payroll_id"`
	UserID        uint64          `gorm:"column:user_id;not null" json:"user_id"`
	WalletID      uint64          `gorm:"column:wallet_id;not null" json:"wallet_id"`
	WalletAddress string          `gorm:"column:wallet_address;not null" json:"wallet_address"`
	WalletType    string          `gorm:"column:wallet_type;not null" json:"wallet_type"`
	WalletChain   string          `gorm:"column:wallet_chain;not null" json:"wallet_chain"`
	Amount        decimal.Decimal `gorm:"column:amount;type:decimal(18,2);not null" json:"amount"`
	Flag          uint8           `gorm:"column:flag" json:"flag"`
	TransTime     time.Time       `gorm:"column:trans_time" json:"trans_time"`
	ReceiptHash   string          `gorm:"column:receipt_hash;not null" json:"receipt_hash"`
}

func (PortalPayslip) TableName() string {
	return "admin_portal_payslip"
}

type PortalSpec struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	SpecName  string    `gorm:"column:spec_name;type:varchar(255);not null" json:"spec_name"`
	SpecValue string    `gorm:"column:spec_value;type:varchar(255);not null" json:"spec_value"`
	SpecType  string    `gorm:"column:spec_type;type:varchar(255);not null" json:"spec_type"`
	AddTime   time.Time `gorm:"column:add_time" json:"add_time"`
	Flag      uint8     `gorm:"column:flag" json:"flag"`
}

func (PortalSpec) TableName() string {
	return "admin_portal_spec"
}
