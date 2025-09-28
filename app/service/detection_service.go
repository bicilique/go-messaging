package service

import (
	"context"
	"go-messaging/model"
)

type DetectionService struct {
	// Add necessary fields here, e.g., repositories, loggers, etc.
}

// NewDetectionService creates a new instance of DetectionService
func NewDetectionService() DetectionInterface {
	return &DetectionService{}
}

// Implement methods for DetectionService here
func (s *DetectionService) SendDetectionNotification(ctx context.Context, request model.DetectionSummary) error {
	// Implement the logic to send detection notification

	// Send to all ?

	return nil
}
