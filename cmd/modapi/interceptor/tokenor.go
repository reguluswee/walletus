package interceptor

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/reguluswee/walletus/cmd/modapi/codes"
	"github.com/reguluswee/walletus/cmd/modapi/common"
	"github.com/reguluswee/walletus/cmd/modapi/security"
	"github.com/reguluswee/walletus/common/log"
	"github.com/reguluswee/walletus/common/model"
	"github.com/reguluswee/walletus/common/system"

	"github.com/gin-gonic/gin"
)

var NoAuthURLs = map[string]bool{
	"/spwapi/admin/portal/login": true,
}

func TokenInterceptor() gin.HandlerFunc {
	return func(c *gin.Context) {
		if NoAuthURLs[c.Request.RequestURI] {
			c.Next()
			return
		}
		allHeaders, ok := c.Get("HEADERS")
		if !ok {
			log.Info("unable to get headers")
			makeFaileRes(c, codes.CODE_ERR_SECURITY, "token check failed")
			return
		}
		allHeader := allHeaders.(common.HeaderParam)

		tokenStr := allHeader.XAuth
		if tokenStr == "" {
			tokenStr = allHeader.AuthToken
		}
		if tokenStr == "" {
			if cookie, err := c.Request.Cookie("AUTH_TOKEN"); err == nil {
				tokenStr = cookie.Value
			}
		}
		token, err := security.Decrypt(tokenStr)
		// log.Info("TOKENCHECK ", tokenStr, token)
		if err != nil {
			makeFaileRes(c, codes.CODE_ERR_SECURITY, "token check failed")
			return
		}
		tokenArr := strings.Split(token, "|")
		if len(tokenArr) != 3 {
			makeFaileRes(c, codes.CODE_ERR_SECURITY, "token length error")
			return
		}
		expireTs, err := strconv.ParseInt(tokenArr[2], 10, 64)
		if err != nil {
			makeFaileRes(c, codes.CODE_ERR_SECURITY, "token format error")
			return
		}
		if time.Now().Unix()-expireTs > int64(common.TOKEN_DURATION.Seconds()) {
			makeFaileRes(c, codes.CODE_ERR_SECURITY, "token expired error")
			return
		}

		mainIdStr := tokenArr[0]
		var db = system.GetDb()
		var portalUser model.PortalUser
		db.Where("id = ?", mainIdStr).Find(&portalUser)
		if portalUser.ID == 0 || portalUser.Flag != 0 {
			makeFaileRes(c, codes.CODE_ERR_SECURITY, "user not existing")
			return
		}

		c.Set("provider_type", tokenArr[1])
		c.Set("main_user", &portalUser)

		c.Next()
	}
}

func makeFaileRes(c *gin.Context, code int64, msg string) {
	c.Abort()
	c.JSON(http.StatusOK, common.Response{
		Code:      code,
		Msg:       msg,
		Timestamp: time.Now().Unix(),
	})
}
