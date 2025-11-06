package model

import (
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/reguluswee/walletus/cmd/modapi/codes"
)

type SysChannel struct {
	ID         uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	AppID      string    `gorm:"column:app_id" json:"app_id"`
	AppKey     string    `gorm:"column:app_key;size:100" json:"app_key"`
	Status     string    `gorm:"column:status" json:"status"`
	Chan       string    `gorm:"column:chan" json:"chan"`
	SigMethod  string    `gorm:"column:sig_method;size:255" json:"sig_method"`
	CreateTime time.Time `gorm:"column:create_time" json:"create_time"`
	UpdateTime time.Time `gorm:"column:update_time" json:"update_time"`
}

func (SysChannel) TableName() string {
	return "sys_channel"
}

func (t *SysChannel) Verify(data, sig string) (bool, int) {
	if t.SigMethod != "SHA256" {
		return false, codes.CODE_ERR_SIGMETHOD_UNSUPP
	}
	if len(data) == 0 || len(sig) == 0 {
		return false, codes.CODE_ERR_AUTHTOKEN_FAIL
	}
	data = fmt.Sprintf("%s%s", data, t.AppKey)

	hashByte := sha256.Sum256([]byte(data))
	hash := fmt.Sprintf("%x", hashByte[:])
	if hash != sig {
		return false, codes.CODE_ERR_AUTHTOKEN_FAIL
	}
	return true, codes.CODE_SUCCESS
}
