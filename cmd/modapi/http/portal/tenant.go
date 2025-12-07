package portal

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/gin-gonic/gin"
	"github.com/reguluswee/walletus/cmd/modapi/codes"
	"github.com/reguluswee/walletus/cmd/modapi/common"
	"github.com/reguluswee/walletus/cmd/modapi/request"
	"github.com/reguluswee/walletus/cmd/modapi/security"
	"github.com/reguluswee/walletus/common/bip"
	"github.com/reguluswee/walletus/common/model"
	"github.com/reguluswee/walletus/common/system"
)

func PortalTenantList(c *gin.Context) {
	res := common.Response{}
	res.Timestamp = time.Now().Unix()

	res.Code = codes.CODE_SUCCESS
	res.Msg = "success"

	mainUser, ok := c.Get("main_user")
	if !ok {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "please login first"
		c.JSON(http.StatusOK, res)
		return
	}
	portalUser, ok := mainUser.(*model.PortalUser)
	if !ok || portalUser == nil {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "please login first"
		c.JSON(http.StatusOK, res)
		return
	}

	var db = system.GetDb()
	var tenants []model.Tenant
	db.Where("flag = ?", 0).Order("add_time DESC").Find(&tenants)
	res.Data = gin.H{
		"tenants": tenants,
	}

	c.JSON(http.StatusOK, res)
}

func PortalTenantCreate(c *gin.Context) {
	var request request.PortalTenantCreateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, common.Response{
			Code:      codes.CODE_ERR_REQFORMAT,
			Msg:       "invalid request: " + err.Error(),
			Timestamp: time.Now().Unix(),
		})
		return
	}

	res := common.Response{
		Timestamp: time.Now().Unix(),
		Code:      codes.CODE_SUCCESS,
		Msg:       "success",
	}

	mainUser, ok := c.Get("main_user")
	if !ok {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "please login first"
		c.JSON(http.StatusOK, res)
		return
	}
	portalUser, ok := mainUser.(*model.PortalUser)
	if !ok || portalUser == nil {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "please login first"
		c.JSON(http.StatusOK, res)
		return
	}

	var db = system.GetDb()

	// 开启事务
	tx := db.Begin()
	if tx.Error != nil {
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "database transaction start failed: " + tx.Error.Error()
		c.JSON(http.StatusOK, res)
		return
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Error("[tenant] create database transaction failed: ", r)
		}
	}()

	appid, appkey := security.GenerateAppIDAndKey(request.UniqueID, request.Name, time.Now().Unix())

	enc, err := bip.GenerateMasterXprv()
	if err != nil {
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "tenant creation error: " + err.Error()
		c.JSON(http.StatusOK, res)
		return
	}

	kdfBytes, _ := json.Marshal(enc.KDF)

	newAPI := model.SysChannel{
		AppID:      appid,
		AppKey:     appkey,
		Status:     "00",
		Chan:       "tenant",
		SigMethod:  "SHA256",
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}

	if err := tx.Create(&newAPI).Error; err != nil {
		tx.Rollback()
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "save api error: " + err.Error()
		c.JSON(http.StatusOK, res)
		return
	}

	newTenant := model.Tenant{
		Name:          request.Name,
		Desc:          request.Desc,
		Callback:      request.Callback,
		Flag:          0,
		AddTime:       time.Now(),
		UniqueID:      request.UniqueID,
		APIID:         newAPI.ID,
		EncMasterXprv: enc.EncMasterXprv,
		EncMasterSeed: enc.EncMasterSeed,
		KdfParams:     string(kdfBytes),
		Version:       "1",
	}

	if err := tx.Create(&newTenant).Error; err != nil {
		tx.Rollback()
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "save tenant error: " + err.Error()
		c.JSON(http.StatusOK, res)
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "database commit failed: " + err.Error()
		c.JSON(http.StatusOK, res)
		return
	}

	res.Data = gin.H{
		"tenant": newTenant,
	}

	c.JSON(http.StatusOK, res)
}

func PortalTenantUpdate(c *gin.Context) {
	var request request.PortalTenantCreateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, common.Response{
			Code:      codes.CODE_ERR_REQFORMAT,
			Msg:       "invalid request: " + err.Error(),
			Timestamp: time.Now().Unix(),
		})
		return
	}

	res := common.Response{
		Timestamp: time.Now().Unix(),
		Code:      codes.CODE_SUCCESS,
		Msg:       "success",
	}

	mainUser, ok := c.Get("main_user")
	if !ok {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "please login first"
		c.JSON(http.StatusOK, res)
		return
	}
	portalUser, ok := mainUser.(*model.PortalUser)
	if !ok || portalUser == nil {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "please login first"
		c.JSON(http.StatusOK, res)
		return
	}

	db := system.GetDb()
	var tenant model.Tenant
	db.Where("id = ? and flag = 0", request.ID).First(&tenant)

	if tenant.ID == 0 {
		res.Code = codes.CODE_ERR_EXIST_OBJ
		res.Msg = "tenant not existing"
		c.JSON(http.StatusOK, res)
		return
	}

	tenant.Name = request.Name
	tenant.Desc = request.Desc
	tenant.Callback = request.Callback
	tenant.UniqueID = request.UniqueID
	db.Save(&tenant)

	res.Data = gin.H{
		"tenant": tenant,
	}

	c.JSON(http.StatusOK, res)
}

func PortalTenantDelete(c *gin.Context) {
	var request request.PortalTenantCreateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, common.Response{
			Code:      codes.CODE_ERR_REQFORMAT,
			Msg:       "invalid request: " + err.Error(),
			Timestamp: time.Now().Unix(),
		})
		return
	}

	res := common.Response{
		Timestamp: time.Now().Unix(),
		Code:      codes.CODE_SUCCESS,
		Msg:       "success",
	}

	mainUser, ok := c.Get("main_user")
	if !ok {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "please login first"
		c.JSON(http.StatusOK, res)
		return
	}
	portalUser, ok := mainUser.(*model.PortalUser)
	if !ok || portalUser == nil {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "please login first"
		c.JSON(http.StatusOK, res)
		return
	}

	db := system.GetDb()
	var tenant model.Tenant
	db.Where("id = ? and flag = 0", request.ID).First(&tenant)

	if tenant.ID == 0 {
		res.Code = codes.CODE_ERR_EXIST_OBJ
		res.Msg = "tenant not existing"
		c.JSON(http.StatusOK, res)
		return
	}

	tenant.Flag = 1
	db.Save(&tenant)

	c.JSON(http.StatusOK, res)
}

func PortalTenantDetail(c *gin.Context) {
	res := common.Response{}
	res.Timestamp = time.Now().Unix()

	res.Code = codes.CODE_SUCCESS
	res.Msg = "success"

	mainUser, ok := c.Get("main_user")
	if !ok {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "please login first"
		c.JSON(http.StatusOK, res)
		return
	}
	portalUser, ok := mainUser.(*model.PortalUser)
	if !ok || portalUser == nil {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "please login first"
		c.JSON(http.StatusOK, res)
		return
	}
	tenantId := c.Param("tenant_id")
	if tenantId == "" {
		res.Code = codes.CODE_ERR_REQFORMAT
		res.Msg = "invalid request: tenant_id is empty"
		c.JSON(http.StatusOK, res)
		return
	}

	var db = system.GetDb()
	var tenant model.Tenant
	var api model.SysChannel
	db.Where("id = ? and flag = 0", tenantId).First(&tenant)
	if tenant.ID == 0 {
		res.Code = codes.CODE_ERR_EXIST_OBJ
		res.Msg = "tenant not existing"
		c.JSON(http.StatusOK, res)
		return
	}
	db.Where("id = ?", tenant.APIID).First(&api)
	res.Data = gin.H{
		"tenant": tenant,
		"api": gin.H{
			"app_id":  api.AppID,
			"app_key": api.AppKey,
		},
	}

	c.JSON(http.StatusOK, res)
}
