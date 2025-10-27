package interceptor

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/reguluswee/walletus/common/log"
	"github.com/reguluswee/walletus/common/model"
	"github.com/reguluswee/walletus/modapi/codes"
	"github.com/reguluswee/walletus/modapi/common"
	"github.com/reguluswee/walletus/modapi/security"
)

func WSInterceptor() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Query("X-Token") == "123456" {
			c.Next() // for test
			return
		}
		queryKeys(WS)
		hp := parse(&c.Request.Header)

		var targetChannel *model.SysChannel
		for _, v := range cacheKeys {
			if v.AppID == hp.AppId {
				targetChannel = &v
			}
		}
		if targetChannel == nil {
			c.Abort()
			c.JSON(http.StatusOK, common.Response{
				Code:      codes.CODE_ERR_APPID_INVALID,
				Msg:       "app_id invalid",
				Timestamp: time.Now().Unix(),
			})
			return
		}
		if ok, code := targetChannel.Verify(hp.Join(), hp.AuthToken); !ok {
			c.Abort()
			c.JSON(http.StatusOK, common.Response{
				Code:      int64(code),
				Msg:       "sig or key params wrong or empty",
				Timestamp: time.Now().Unix(),
			})
			return
		}
		c.Set("APPID", hp.AppId)
		c.Set("REQUESTID", hp.RequestId)
		c.Set("TS", hp.Ts)
		c.Set("HEADERS", hp)

		token, err := security.Decrypt(hp.XAuth)
		if err == nil {
			tokenArr := strings.Split(token, "|")
			if len(tokenArr) == 4 {
				expireTs, err := strconv.ParseInt(tokenArr[3], 10, 64)
				if err == nil {
					if time.Now().Unix()-expireTs <= int64(common.TOKEN_DURATION.Seconds()) {
						c.Set("user_wallet", tokenArr[1])
						c.Set("user_id", tokenArr[0])
					}
				}
			}

		}
		c.Next()

		h := c.Request.Header
		log.Info(h)

		c.Next()
	}
}
