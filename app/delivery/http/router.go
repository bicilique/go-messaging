package http

import "github.com/gin-gonic/gin"

type RouteConfig struct {
	Router *gin.Engine
	// IrisHandler *IrisHandler
	UserHandler *UserHandler
}

func (c *RouteConfig) Setup() {
	// Iris webhook routes
	// c.Router.POST("/iris/send-message", c.IrisHandler.SendTelegramMessage)
	// c.Router.POST("/iris/send-notification", c.IrisHandler.SendTelegramNotification)

	// API v1 routes
	v1 := c.Router.Group("/api/v1")
	{
		// User routes
		users := v1.Group("/users")
		{
			users.POST("", c.UserHandler.CreateUser)
			users.GET("", c.UserHandler.ListUsers)
			users.GET("/:id", c.UserHandler.GetUser)
			users.PUT("/:id", c.UserHandler.UpdateUser)
			users.GET("/telegram/:telegram_user_id", c.UserHandler.GetUserByTelegramID)
			users.DELETE("/telegram/:telegram_user_id", c.UserHandler.DeleteUser)
		}
	}
}
