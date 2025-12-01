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

	adminGroup.GET("/portal/rbac/user/menus", portal.PortalUserMenus)

	adminGroup.GET("/portal/rbac/role/list", portal.PortalRoleList)
	adminGroup.POST("/portal/rbac/role/create", portal.PortalRoleCreate)
	adminGroup.POST("/portal/rbac/role/update", portal.PortalRoleUpdate)
	adminGroup.POST("/portal/rbac/role/delete", portal.PortalRoleDelete)
	adminGroup.GET("/portal/rbac/func/list", portal.PortalFuncList)

	adminGroup.GET("/portal/rbac/role/func/list/:role_id", portal.PortalRoleFuncList)
	adminGroup.GET("/portal/rbac/role/user/list/:role_id", portal.PortalRoleUserList)
	adminGroup.POST("/portal/rbac/role/permission/func/bind/:role_id/:func_id", portal.PortalRoleFuncBind)
	adminGroup.POST("/portal/rbac/role/permission/user/bind/:role_id/:user_id", portal.PortalRoleUserBind)
	adminGroup.POST("/portal/rbac/role/permission/func/unbind/:role_id/:func_id", portal.PortalRoleFuncUnbind)
	adminGroup.POST("/portal/rbac/role/permission/user/unbind/:role_id/:user_id", portal.PortalRoleUserUnbind)

	adminGroup.GET("/portal/dashboard", portal.PortalDashboard)
	adminGroup.GET("/portal/dept/list", portal.PortalDeptList)
	adminGroup.POST("/portal/dept/create", portal.PortalDeptCreate)
	adminGroup.POST("/portal/dept/update", portal.PortalDeptUpdate)
	adminGroup.POST("/portal/dept/delete", portal.PortalDeptDelete)
	adminGroup.GET("/portal/user/list", portal.PortalUserList)
	adminGroup.POST("/portal/user/update", portal.PortalUserUpdate)

	adminGroup.GET("/portal/payroll/list", portal.PortalPayrollList)
	adminGroup.POST("/portal/payroll/create", portal.PortalPayrollCreate)
	adminGroup.POST("/portal/payroll/update", portal.PortalPayrollUpdate)
	adminGroup.POST("/portal/payroll/delete", portal.PortalPayrollDelete)
	adminGroup.GET("/portal/payroll/staff/list", portal.PortalPayrollStaffList)
	adminGroup.POST("/portal/payroll/staff/wallet/:user_id", portal.PortalPayrollStaffWallet)
	adminGroup.GET("/portal/payroll/detail/:payroll_id", portal.PortalPayrollDetail)

}
