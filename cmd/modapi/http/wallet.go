package http

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/reguluswee/walletus/cmd/modapi/common"
	"github.com/reguluswee/walletus/cmd/modapi/request"
	"github.com/reguluswee/walletus/cmd/modapi/service"
	"github.com/reguluswee/walletus/common/bip"
	"github.com/reguluswee/walletus/common/chain"
	"github.com/reguluswee/walletus/common/chain/dep"
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
		return
	}
	res := common.Response{}
	res.Timestamp = time.Now().Unix()

	res.Code = codes.CODE_SUCCESS
	res.Msg = "success"

	var db = system.GetDb()
	tenantId, exist := c.Get("TENANTID")
	if !exist {
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "tenant not existed"
		c.JSON(http.StatusOK, res)
		return
	}

	var tenant model.Tenant
	db.Where("id = ?", tenantId).First(&tenant)
	if tenant.ID == 0 {
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "tenant not found"
		c.JSON(http.StatusOK, res)
		return
	}

	tenantAddrId, addr, err := service.WalletCreate(request, tenant)
	if err != nil {
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = err.Error()
		c.JSON(http.StatusOK, res)
		return
	}

	res.Data = gin.H{
		"address_id": tenantAddrId,
		"address":    addr,
		"chain":      request.Chain,
	}

	c.JSON(http.StatusOK, res)
}

func WalletBalanceQuery(c *gin.Context) {
	var request request.WalletBalanceQueryRequest
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

	db := system.GetDb()
	var tenant model.Tenant
	db.Where("id = ?", request.TenantID).First(&tenant)
	if tenant.ID == 0 {
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "tenant not existed"
		c.JSON(http.StatusOK, res)
		return
	}
	var tenantAddress model.TenantAddress
	db.Where("tenant_id = ? and address_index = ?", tenant.ID, request.AddressID).First(&tenantAddress)
	if tenantAddress.ID == 0 {
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "tenant address not existed"
		c.JSON(http.StatusOK, res)
		return
	}

	chainDef, err := bip.CheckValidChainCode(request.Chain)
	if err != nil {
		res.Code = codes.CODE_ERR_METHOD_UNSUPPORT
		res.Msg = err.Error()
		c.JSON(http.StatusOK, res)
		return
	}

	gw := chain.NewGateway()
	q := chain.BalanceQuery{
		Chain:     chainDef,
		Network:   "mainnet",
		Addresses: []string{tenantAddress.AddressVal},
		Tokens: map[string][]string{
			tenantAddress.AddressVal: {request.Token},
		},
		Consistency: dep.Consistency{Mode: "safe", MinConfirmations: 0},
	}
	responseValue, err := gw.GetBalances(context.Background(), q)
	if err != nil {
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = err.Error()
		c.JSON(http.StatusOK, res)
		return
	}

	res.Code = codes.CODE_SUCCESS
	res.Data = gin.H{
		"data": responseValue,
	}
	res.Msg = "success"
	c.JSON(http.StatusOK, res)
}
