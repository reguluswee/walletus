package http

import (
	"github.com/gin-gonic/gin"
)

func Routers(e *gin.RouterGroup) {

	homeGroup := e.Group("/")
	homeGroup.GET("public", Public)

}
