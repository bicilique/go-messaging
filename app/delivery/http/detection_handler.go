package http

import (
	"go-messaging/model"
	"go-messaging/service"

	"github.com/gin-gonic/gin"
)

type DetectionHandler struct {
	detectionService service.DetectionInterface
}

// NewDetectionHandler creates a new instance of DetectionHandler
func NewDetectionHandler(detectionService service.DetectionInterface) *DetectionHandler {
	return &DetectionHandler{
		detectionService: detectionService,
	}
}

// SendDetectionNotification handles the HTTP request to send detection notifications
func (h *DetectionHandler) SendDetectionNotification(c *gin.Context) {
	var req model.DetectionSummary
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request payload"})
		return
	}
	if err := h.detectionService.SendDetectionNotification(c.Request.Context(), req); err != nil {
		c.JSON(500, gin.H{"error": "Failed to send detection notification"})
		return
	}
	c.JSON(200, gin.H{"message": "Detection notification sent successfully"})
}
