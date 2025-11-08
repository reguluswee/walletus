package router

import (
	"github.com/gin-gonic/gin"
	"github.com/reguluswee/walletus/cmd/modapi/http"
)

func SubRouters(e *gin.RouterGroup) {

	homeGroup := e.Group("/")
	homeGroup.GET("public", http.Public)

	homeGroup.POST("/tenant/create", http.TenantCreate)
	homeGroup.POST("/was/create", http.WalletCreate)

}
