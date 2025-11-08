package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/reguluswee/walletus/cmd/modapi/codes"
	"github.com/reguluswee/walletus/cmd/modapi/common"
	"github.com/reguluswee/walletus/common/bip"
)

func Public(c *gin.Context) {
	res := common.Response{}
	res.Timestamp = time.Now().Unix()

	res.Code = codes.CODE_SUCCESS
	res.Msg = "success"

	res.Data = gin.H{
		"supported_chains": bip.SupportChains(),
	}

	c.JSON(http.StatusOK, res)
}
