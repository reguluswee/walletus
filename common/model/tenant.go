package model

import "time"

type Tenant struct {
	ID            uint64    `gorm:"primaryKey;autoIncrement"`
	UniqueID      string    `gorm:"column:unique_id;type:varchar(100);not null" json:"unique_id"`
	Name          string    `gorm:"column:name;type:varchar(255);not null" json:"name"`
	EncMasterXprv string    `gorm:"column:enc_master_xprv;type:varchar(255);not null"`
	KdfParams     string    `gorm:"column:kdf_params;type:varchar(255);not null"`
	AddTime       time.Time `gorm:"column:add_time" json:"add_time"`
	Version       string    `gorm:"column:version;type:varchar(255);not null" json:"version"`
	Callback      string    `gorm:"column:call_back;type:varchar(255);not null" json:"call_back"`
}

func (Tenant) TableName() string {
	return "tenant_information"
}

type TenantChain struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement"`
	TenantID    uint64    `gorm:"column:tenant_id;not null"`
	Chain       string    `gorm:"column:chain;type:varchar(255);not null"`
	CoinType    uint32    `gorm:"column:coin_type;type:int(11);not null"`
	XPub        string    `gorm:"column:x_pub;type:varchar(100);not null" json:"x_pub"`
	DerivedPath string    `gorm:"column:derived_path;type:varchar(100);not null" json:"derived_path"`
	AddTime     time.Time `gorm:"column:add_time" json:"add_time"`
}

func (TenantChain) TableName() string {
	return "tenant_chain"
}

type TenantAddress struct {
	ID            uint64    `gorm:"primaryKey;autoIncrement"`
	TenantID      uint64    `gorm:"column:tenant_id;type:int(11);not null"`
	TenantChainID uint64    `gorm:"column:tenant_chain_id;not null"`
	AddressIndex  uint32    `gorm:"column:address_index;not null"`
	AddressVal    string    `gorm:"column:address_val;not null"`
	DerivedPath   string    `gorm:"column:derived_path;type:varchar(100);not null" json:"derived_path"`
	AddTime       time.Time `gorm:"column:add_time" json:"add_time"`
}

func (TenantAddress) TableName() string {
	return "tenant_address"
}
