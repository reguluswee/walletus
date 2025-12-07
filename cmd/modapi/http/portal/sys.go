package portal

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/reguluswee/walletus/cmd/modapi/codes"
	"github.com/reguluswee/walletus/cmd/modapi/common"
	"github.com/reguluswee/walletus/common/model"
	"github.com/reguluswee/walletus/common/system"
)

const (
	SPEC_TYPE_PAYROLL_SETTINGS = "payroll_settings"
)

type PayrollSettings struct {
	Chain       string `json:"chain"`
	PayContract string `json:"pay_contract"`
	PayToken    string `json:"pay_token"`
}

func (t PayrollSettings) IsValid() bool {
	if t.Chain == "" || t.PayContract == "" || t.PayToken == "" {
		return false
	}
	return true
}

func PortalPayrollSettings(c *gin.Context) {
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
	var portalSpecs []model.PortalSpec
	db.Where("flag = ? and spec_type = ?", 0, SPEC_TYPE_PAYROLL_SETTINGS).Find(&portalSpecs)

	var result PayrollSettings
	for _, spec := range portalSpecs {
		switch spec.SpecName {
		case "chain":
			result.Chain = spec.SpecValue
		case "pay_contract":
			result.PayContract = spec.SpecValue
		case "pay_token":
			result.PayToken = spec.SpecValue
		}
	}
	var configMap = map[string]interface{}{
		"arbitrum": map[string]string{
			"usdt": "0xFd086bC7CD5C481DCC9C85ebE478A1C0b69FCbb9",
			"usdc": "0xaf88d065e77c8cC2239327C5EDb3A432268e5831",
		},
	}
	res.Data = gin.H{
		"payroll_settings": result,
		"config_map":       configMap,
	}

	c.JSON(http.StatusOK, res)
}

func PortalPayrollSettingsSave(c *gin.Context) {
	var request PayrollSettings
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, common.Response{
			Code:      codes.CODE_ERR_REQFORMAT,
			Msg:       "invalid request: " + err.Error(),
			Timestamp: time.Now().Unix(),
		})
		return
	}
	res := common.Response{}
	res.Timestamp = time.Now().Unix()

	var db = system.GetDb()
	db.Where("flag = ? and spec_type = ?", 0, SPEC_TYPE_PAYROLL_SETTINGS).Delete(&model.PortalSpec{})

	var portalSpec []model.PortalSpec
	if request.Chain != "" {
		portalSpec = append(portalSpec, model.PortalSpec{
			SpecName:  "chain",
			SpecValue: request.Chain,
			SpecType:  SPEC_TYPE_PAYROLL_SETTINGS,
			AddTime:   time.Now(),
			Flag:      0,
		})
	}
	if request.PayContract != "" {
		portalSpec = append(portalSpec, model.PortalSpec{
			SpecName:  "pay_contract",
			SpecValue: request.PayContract,
			SpecType:  SPEC_TYPE_PAYROLL_SETTINGS,
			AddTime:   time.Now(),
			Flag:      0,
		})
	}
	if request.PayToken != "" {
		portalSpec = append(portalSpec, model.PortalSpec{
			SpecName:  "pay_token",
			SpecValue: request.PayToken,
			SpecType:  SPEC_TYPE_PAYROLL_SETTINGS,
			AddTime:   time.Now(),
			Flag:      0,
		})
	}
	if len(portalSpec) > 0 {
		err := db.Create(&portalSpec).Error
		if err != nil {
			res.Code = codes.CODE_ERR_UNKNOWN
			res.Msg = "database error: " + err.Error()
			c.JSON(http.StatusOK, res)
			return
		}
	}

	c.JSON(http.StatusOK, res)
}
