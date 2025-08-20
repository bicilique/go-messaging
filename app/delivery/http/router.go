package http

import "github.com/gin-gonic/gin"

type RouteConfig struct {
	Router      *gin.Engine
	IrisHandler *IrisHandler
}

func (c *RouteConfig) Setup() {
	c.Router.POST("/iris/send-message", c.IrisHandler.SendTelegramMessage)
	c.Router.POST("/iris/send-notification", c.IrisHandler.SendTelegramNotification)
}
