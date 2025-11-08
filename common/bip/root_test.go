package bip

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/reguluswee/walletus/common/model"
	"github.com/reguluswee/walletus/common/system"
)

func TestGenerateMasterXprv(t *testing.T) {

	tenant := model.Tenant{
		Name: "test-proj",
	}
	db := system.GetDb()
	enc, _ := GenerateMasterXprv()

	kdfBytes, _ := json.Marshal(enc.KDF)

	tenant.AddTime = time.Now()
	tenant.EncMasterXprv = enc.EncMasterXprv
	tenant.KdfParams = string(kdfBytes)
	tenant.Version = "1"
	db.Save(&tenant)

	chainxpubETH, _ := GenerateDerivationChain(uint32(tenant.ID), enc, "ETH")
	tenantChainETH := model.TenantChain{
		TenantID:    tenant.ID,
		Chain:       "ETH",
		CoinType:    60,
		XPub:        chainxpubETH.XPub,
		DerivedPath: chainxpubETH.DerivedPath,
	}
	err := db.Save(&tenantChainETH).Error
	fmt.Println(err)
	i := 1
	for i <= 2 {
		addr, path, _ := DeriveAddressFromXpub(chainxpubETH.XPub, uint32(tenant.ID), uint32(i), "ETH")
		db.Save(&model.TenantAddress{
			TenantID:      tenant.ID,
			TenantChainID: tenantChainETH.ID,
			AddressIndex:  uint32(i),
			AddressVal:    addr,
			DerivedPath:   path,
		})
		i++
	}

	chainxpubTRON, _ := GenerateDerivationChain(uint32(tenant.ID), enc, "TRON")
	tenantChainTRON := model.TenantChain{
		TenantID:    tenant.ID,
		Chain:       "TRON",
		CoinType:    195,
		XPub:        chainxpubTRON.XPub,
		DerivedPath: chainxpubTRON.DerivedPath,
	}
	err = db.Save(&tenantChainTRON).Error
	fmt.Println(err)
	i = 1
	for i <= 2 {
		addr, path, _ := DeriveAddressFromXpub(chainxpubTRON.XPub, uint32(tenant.ID), uint32(i), "TRON")
		db.Save(&model.TenantAddress{
			TenantID:      tenant.ID,
			TenantChainID: tenantChainTRON.ID,
			AddressIndex:  uint32(i),
			AddressVal:    addr,
			DerivedPath:   path,
		})
		i++
	}
}

func TestDerivedPri(t *testing.T) {
	db := system.GetDb()
	var tenant model.Tenant
	db.Where("id = 1").First(&tenant)

	var kdf KDFParams
	json.Unmarshal([]byte(tenant.KdfParams), &kdf)
	em := EncMaster{
		EncMasterXprv: tenant.EncMasterXprv,
		KDF:           kdf,
	}

	exkey, err := DeriveChildXprv(em, "m/44'/60'/1'/0/2")
	if err != nil {
		fmt.Println(err)
	} else {
		privatekey := exkey.String()
		fmt.Println(privatekey)
	}

	addr, pri, err := AddressAndPrivFromPath(em, "m/44'/60'/1'/0/2", "ETH")
	fmt.Println(addr, pri)
}
