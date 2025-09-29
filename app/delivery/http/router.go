package http

import "github.com/gin-gonic/gin"

type RouteConfig struct {
	Router           *gin.Engine
	UserHandler      *UserHandler
	AdminHandler     *AdminHandler
	AuthMiddleware   *BasicAuthMiddleware
	DetectionHandler *DetectionHandler
}

func (c *RouteConfig) Setup() {
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

		// Admin routes with authentication
		if c.AdminHandler != nil {
			admin := v1.Group("/admin")

			// Apply authentication middleware
			if c.AuthMiddleware != nil {
				admin.Use(c.AuthMiddleware.AdminAuth())
			} else {
				// Fallback to simple basic auth for development
				admin.Use(SimpleBasicAuth("admin", "admin123"))
			}

			{
				admin.POST("/create", c.AdminHandler.CreateAdmin)
				admin.GET("/users/pending", c.AdminHandler.GetPendingUsers)
				admin.GET("/users/approved", c.AdminHandler.GetApprovedUsers)
				admin.POST("/users/:userID/approve", c.AdminHandler.ApproveUser)
				admin.POST("/users/:userID/reject", c.AdminHandler.RejectUser)
				admin.POST("/users/:userID/disable", c.AdminHandler.DisableUser)
				admin.POST("/users/:userID/enable", c.AdminHandler.EnableUser)
				admin.GET("/stats", c.AdminHandler.GetUserStats)
				admin.POST("/cleanup", c.AdminHandler.CleanupPendingUsers)
			}
		}

		// Detection routes
		if c.DetectionHandler != nil {
			detection := v1.Group("/detection")
			{
				detection.POST("/notify", c.DetectionHandler.SendDetectionNotification)
			}
		}
	}
}
