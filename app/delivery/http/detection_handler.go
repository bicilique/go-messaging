package http

import (
	"go-messaging/model"

	"github.com/gin-gonic/gin"
)

type DetectionHandler struct {
	// Add necessary fields here
}

// NewDetectionHandler creates a new instance of DetectionHandler
func NewDetectionHandler() *DetectionHandler {
	return &DetectionHandler{}
}

// Add methods for DetectionHandler here
func (h *DetectionHandler) SendDetectionNotification(c *gin.Context) {
	var req model.DetectionSummary
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request payload"})
		return
	}

	c.JSON(200, gin.H{"status": "Detection notification received", "data": req})
}
