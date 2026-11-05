package wxcallback

import (
	"github.com/gin-gonic/gin"
)

// Routers 路由
func Routers(e *gin.Engine) {
	// WeChat calls these public callback URLs directly and does not include the
	// platform-specific x-wx-source header checked by WXSourceMiddleWare.
	// Keep callbacks reachable so their payloads can be persisted and handled.
	g := e.Group("/wxcallback")
	g.GET("/component", componentHandler)
	g.POST("/component", componentHandler)
	g.POST("/biz/:appid", bizHandler)
	g.GET("/biz/:appid", bizHandler)
}
