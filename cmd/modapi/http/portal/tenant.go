package portal

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/reguluswee/walletus/cmd/modapi/codes"
	"github.com/reguluswee/walletus/cmd/modapi/common"
	"github.com/reguluswee/walletus/cmd/modapi/request"
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

	newTenant := model.Tenant{
		Name:     request.Name,
		Desc:     request.Desc,
		Callback: request.Callback,
		Flag:     0,
		AddTime:  time.Now(),
	}
	db := system.GetDb()
	db.Create(&newTenant)

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
