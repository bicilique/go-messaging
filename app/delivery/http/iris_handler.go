package http

import (
	"go-messaging/model"
	"go-messaging/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type IrisHandler struct {
	TelegramService service.TelegramService
}

func NewIrisHandler(telegramService service.TelegramService) *IrisHandler {
	return &IrisHandler{
		TelegramService: telegramService,
	}
}

func (h *IrisHandler) SendTelegramMessage(c *gin.Context) {
	// Call the Telegram service to send a message
	var req model.SendMessageRequest

	// Bind the incoming JSON request to the SendMessageRequest struct
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request payload",
			"details": err.Error(),
		})
		return
	}

	// Call the Telegram service to send a message using the helper method for string chat IDs
	err := h.TelegramService.SendMessageByStringID(req.ChatID, req.Message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to send message",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Message sent successfully",
		"chat_id": req.ChatID,
	})

}

func (h *IrisHandler) SendTelegramNotification(c *gin.Context) {
	var req model.IocPayload

	// Bind the incoming JSON request to the IocPayload struct
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request payload",
			"details": err.Error(),
		})
		return
	}

	err := h.TelegramService.SendIocMessage("-4699507254", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to send notification",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":      "success",
		"message":     "Notification sent successfully",
		"case_id":     req.CaseID,
		"link":        req.Link,
		"description": req.Description,
	})
}

// func (h *IrisHandler) SendTelegramNotification(c *gin.Context) {
// 	// This handler can be used to send notifications, similar to SendTelegramMessage
// 	bodyBytes, err := c.GetRawData()
// 	if err != nil {
// 		log.Printf("Error reading request body: %v", err)
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Could not read request body"})
// 		return
// 	}

// 	// Log the raw body for debugging
// 	log.Printf("--- Received Raw Webhook Body ---\n%s\n---------------------------------", string(bodyBytes))

// 	var payload model.WebhookPayload
// 	if err := json.Unmarshal(bodyBytes, &payload); err != nil {
// 		log.Printf("Error unmarshaling JSON: %v", err)
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
// 		return
// 	}

// 	// 4. Now you can work with the structured data.
// 	log.Println("✅ Successfully unmarshaled payload.")
// 	if len(payload.Embeds) > 0 {
// 		log.Printf("Processing Case: %s", payload.Embeds[0].Title)
// 	}

// 	// 5. Send a success response back to Iris.
// 	c.JSON(http.StatusOK, gin.H{"status": "notification received and processed"})
// }
