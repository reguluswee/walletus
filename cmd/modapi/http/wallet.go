package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/reguluswee/walletus/cmd/modapi/common"
	"github.com/reguluswee/walletus/cmd/modapi/request"
	"github.com/reguluswee/walletus/cmd/modapi/service"
	"github.com/reguluswee/walletus/common/codes"
	"github.com/reguluswee/walletus/common/model"
	"github.com/reguluswee/walletus/common/system"
)

func WalletCreate(c *gin.Context) {
	var request request.WalletCreateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, common.Response{
			Code:      codes.CODE_ERR_REQFORMAT,
			Msg:       "invalid request",
			Data:      nil,
			Timestamp: time.Now().Unix(),
		})
	}
	res := common.Response{}
	res.Timestamp = time.Now().Unix()

	res.Code = codes.CODE_SUCCESS
	res.Msg = "success"

	var db = system.GetDb()
	var tenant model.Tenant
	db.Where("id = ?", request.TenantID).First(&tenant)
	if tenant.ID == 0 {
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "tenant not found"
		c.JSON(http.StatusOK, res)
		return
	}

	addr, err := service.WalletCreate(request, tenant)
	if err != nil {
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = err.Error()
		c.JSON(http.StatusOK, res)
		return
	}

	res.Data = gin.H{
		"rpc": map[string]string{
			"address": addr,
			"chain":   request.Chain,
		},
	}

	c.JSON(http.StatusOK, res)
}
