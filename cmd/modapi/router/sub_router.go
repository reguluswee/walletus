package router

import (
	"github.com/gin-gonic/gin"
	"github.com/reguluswee/walletus/cmd/modapi/http"
	"github.com/reguluswee/walletus/cmd/modapi/http/portal"
	"github.com/reguluswee/walletus/cmd/modapi/interceptor"
)

func SubRouters(e *gin.RouterGroup) {

	homeGroup := e.Group("/")
	homeGroup.GET("public", http.Public)

	homeGroup.POST("/tenant/create", http.TenantCreate)
	homeGroup.POST("/tenant/update", http.TenantUpdate)

	homeGroup.POST("/was/create", http.WalletCreate)
	homeGroup.POST("/was/balance/query", http.WalletBalanceQuery)

	adminGroup := e.Group("/admin", interceptor.TokenInterceptor())
	adminGroup.POST("/portal/login", portal.PortalLogin)
	adminGroup.GET("/portal/dashboard", portal.PortalDashboard)
	adminGroup.GET("/portal/dept/list", portal.PortalDeptList)
	adminGroup.POST("/portal/dept/create", portal.PortalDeptCreate)
	adminGroup.POST("/portal/dept/update", portal.PortalDeptUpdate)
	adminGroup.POST("/portal/dept/delete", portal.PortalDeptDelete)
	adminGroup.GET("/portal/user/list", portal.PortalUserList)
	adminGroup.POST("/portal/user/update", portal.PortalUserUpdate)
}
