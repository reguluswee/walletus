package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/reguluswee/walletus/cmd/modapi/common"
	"github.com/reguluswee/walletus/cmd/modapi/request"
	"github.com/reguluswee/walletus/common/bip"
	"github.com/reguluswee/walletus/common/codes"
	"github.com/reguluswee/walletus/common/model"
	"github.com/reguluswee/walletus/common/system"
)

func TenantCreate(c *gin.Context) {
	var request request.TenantCreateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, common.Response{
			Code:      codes.CODE_ERR_REQFORMAT,
			Msg:       "invalid request",
			Data:      nil,
			Timestamp: time.Now().Unix(),
		})
		return
	}
	res := common.Response{}
	res.Timestamp = time.Now().Unix()

	res.Code = codes.CODE_SUCCESS
	res.Msg = "success"

	var db = system.GetDb()
	var tenant model.Tenant
	db.Where("unique_id = ?", request.UniqueID).First(&tenant)
	if tenant.ID > 0 {
		res.Code = codes.CODE_ERR_EXIST_OBJ
		res.Msg = "tenant unique id existed"
		c.JSON(http.StatusOK, res)
		return
	}

	enc, err := bip.GenerateMasterXprv()
	if err != nil {
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "tenant creation error: " + err.Error()
		c.JSON(http.StatusOK, res)
		return
	}

	kdfBytes, _ := json.Marshal(enc.KDF)

	tenant = model.Tenant{
		Name:          request.Name,
		UniqueID:      request.UniqueID,
		AddTime:       time.Now(),
		EncMasterXprv: enc.EncMasterXprv,
		EncMasterSeed: enc.EncMasterSeed,
		KdfParams:     string(kdfBytes),
		Version:       "1",
		Callback:      request.Callback,
	}

	db.Save(&tenant)

	res.Data = gin.H{
		"tenant_id":        tenant.ID,
		"tenant_unique_id": tenant.UniqueID,
	}

	c.JSON(http.StatusOK, res)
}

func TenantUpdate(c *gin.Context) {
	var request request.TenantCreateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, common.Response{
			Code:      codes.CODE_ERR_REQFORMAT,
			Msg:       "invalid request",
			Data:      nil,
			Timestamp: time.Now().Unix(),
		})
		return
	}
	res := common.Response{}
	res.Timestamp = time.Now().Unix()

	res.Code = codes.CODE_SUCCESS
	res.Msg = "success"

	var db = system.GetDb()
	var tenant model.Tenant
	db.Where("unique_id = ?", request.UniqueID).First(&tenant)
	if tenant.ID == 0 {
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "tenant not existed"
		c.JSON(http.StatusOK, res)
		return
	}

	tenant.Name = request.Name
	tenant.Callback = request.Callback
	db.Save(&tenant)

	res.Data = gin.H{
		"tenant_id":        tenant.ID,
		"tenant_unique_id": tenant.UniqueID,
	}

	c.JSON(http.StatusOK, res)
}
