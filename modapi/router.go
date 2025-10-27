package router

import (
	"net/http"
	"os"

	general "github.com/reguluswee/walletus/modapi/http"
	"github.com/reguluswee/walletus/modapi/interceptor"

	"github.com/reguluswee/walletus/common/log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Option func(*gin.RouterGroup)

var options = []Option{}

func Include(opts ...Option) {
	options = append(options, opts...)
}

func Init() *gin.Engine {
	Include(general.Routers)

	r := gin.New()

	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	env := os.Getenv("ENV")

	log.Info(env)
	if env == "dev" {
		r.Use(cors.New(cors.Config{
			AllowOrigins:     []string{"*"},
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "APPID", "SIG", "TS", "VER", "REQUESTID", "XAUTH", "DAUTH"},
			ExposeHeaders:    []string{"Content-Length"},
			AllowCredentials: true,
		}))
	}

	r.GET("/index", helloHandler) //Default welcome api

	// wsGroup := r.Group("/ws", interceptor.WSInterceptor())
	// wsGroup.GET("chat", ws.Chat)

	// Twitter OAuth login endpoint - no authentication required
	// r.GET("/spwapi/preauth/thirdpart/x/login", auth.XLogin)

	apiGroup := r.Group("/spwapi", interceptor.HttpInterceptor()) // total interceptor stack
	for _, opt := range options {
		opt(apiGroup)
	}

	return r
}

func helloHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Hello World",
	})
}
