package service

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/reguluswee/walletus/cmd/modapi/request"
	"github.com/reguluswee/walletus/common/bip"
	"github.com/reguluswee/walletus/common/model"
	"github.com/reguluswee/walletus/common/system"
	"gorm.io/gorm"
)

func WalletCreate(request request.WalletCreateRequest, tenant model.Tenant) (uint64, string, error) {
	chainDef, err := bip.CheckValidChainCode(request.Chain)
	if err != nil {
		return 0, "", errors.New("unsupport chain:" + request.Chain)
	}
	var kdf bip.KDFParams
	err = json.Unmarshal([]byte(tenant.KdfParams), &kdf)
	if err != nil {
		return 0, "", errors.New("tenant kdf params error:" + err.Error())
	}
	enc := bip.EncMaster{
		EncMasterXprv: tenant.EncMasterXprv,
		KDF:           kdf,
	}

	committed := false
	var db = system.GetDb()
	tx := db.Begin()

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			return
		}
		if !committed {
			_ = tx.Rollback()
		}
	}()

	var tenantChain model.TenantChain
	if err := tx.Where("tenant_id = ? and chain = ?", request.TenantID, request.Chain).First(&tenantChain).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, "", errors.New("unknown error:" + err.Error())
		}
		chainDerivedPath, err := bip.GenerateDerivationChain(uint32(tenant.ID), enc, chainDef.Name)
		if err != nil {
			return 0, "", errors.New("generate derivation chain error:" + err.Error())
		}
		tenantChain = model.TenantChain{
			TenantID:    tenant.ID,
			Chain:       chainDef.Name,
			CoinType:    chainDef.CoinType,
			XPub:        chainDerivedPath.XPub,
			DerivedPath: chainDerivedPath.DerivedPath,
			AddTime:     time.Now(),
		}
		if err := tx.Save(&tenantChain).Error; err != nil {
			return 0, "", errors.New("save tenant chain error:" + err.Error())
		}
	}

	var tenantAddress model.TenantAddress
	tx.Where("tenant_id = ? and tenant_chain_id = ? and address_index = ?", tenant.ID, tenantChain.ID, request.UniqueID).First(&tenantAddress)
	if tenantAddress.ID != 0 {
		committed = true
		return tenantAddress.ID, tenantAddress.AddressVal, nil
	}

	addr, path, err := bip.DeriveAddressFromXpub(tenantChain.XPub, uint32(tenant.ID), request.UniqueID, tenantChain.Chain)

	if err != nil {
		return 0, "", errors.New("derive address from xpub error:" + err.Error())
	}

	tenantAddress = model.TenantAddress{
		TenantID:      tenant.ID,
		TenantChainID: tenantChain.ID,
		AddressIndex:  request.UniqueID,
		AddressVal:    addr,
		DerivedPath:   path,
		AddTime:       time.Now(),
	}
	if err := tx.Save(&tenantAddress).Error; err != nil {
		return 0, "", errors.New("save tenant address error:" + err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		return 0, "", errors.New("commit transaction error:" + err.Error())
	}
	committed = true

	return tenantAddress.ID, addr, nil
}
