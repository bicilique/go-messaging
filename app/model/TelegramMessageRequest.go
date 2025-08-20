package model

type SendMessageRequest struct {
	ChatID  string `json:"chat_id" binding:"required"`
	Message string `json:"message" binding:"required"`
}
