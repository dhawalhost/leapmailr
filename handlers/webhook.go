package handlers

import (
	"net/http"
	"time"

	"github.com/dhawalhost/leapmailr/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// WebhookHandler processes email provider webhooks (SendGrid, Mailgun, etc.)
func WebhookHandler(c *gin.Context) {
	provider := c.Param("provider")

	switch provider {
	case "sendgrid":
		handleSendGridWebhook(c)
	case "mailgun":
		handleMailgunWebhook(c)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported provider"})
	}
}

// SendGrid webhook payload structure
type SendGridEvent struct {
	Email     string `json:"email"`
	Event     string `json:"event"`
	Reason    string `json:"reason"`
	Status    string `json:"status"`
	Timestamp int64  `json:"timestamp"`
	UserID    string `json:"userid"` // Custom field we can add to track user
}

func handleSendGridWebhook(c *gin.Context) {
	var events []SendGridEvent
	if err := c.ShouldBindJSON(&events); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	trackingService := service.NewWebhookTrackingService()
	processedCount := 0

	for _, event := range events {
		// Parse user ID if provided
		var userID *uuid.UUID
		if event.UserID != "" {
			if uid, err := uuid.Parse(event.UserID); err == nil {
				userID = &uid
			}
		}

		// Create metadata
		metadata := map[string]interface{}{
			"event":     event.Event,
			"reason":    event.Reason,
			"status":    event.Status,
			"timestamp": event.Timestamp,
			"provider":  "sendgrid",
		}

		// Process webhook event to update email status
		webhookEvent := service.EmailEvent{
			MessageID: "", // SendGrid may include sg_message_id in custom args
			Email:     event.Email,
			Event:     event.Event,
			Reason:    event.Reason,
			Timestamp: parseTimestamp(event.Timestamp),
			Provider:  "sendgrid",
			Metadata:  metadata,
			UserID:    userID,
		}

		if err := trackingService.ProcessWebhookEvent(webhookEvent); err != nil {
			// Log error but continue processing other events
			continue
		}

		processedCount++
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "processed",
		"processed": processedCount,
		"total":     len(events),
	})
}

// parseTimestamp converts Unix timestamp to time.Time
func parseTimestamp(timestamp int64) time.Time {
	if timestamp == 0 {
		return time.Now()
	}
	return time.Unix(timestamp, 0)
}

// Mailgun webhook payload structure
type MailgunWebhookPayload struct {
	EventData MailgunEventData `json:"event-data"`
}

type MailgunEventData struct {
	Event     string                 `json:"event"`
	Recipient string                 `json:"recipient"`
	Reason    string                 `json:"reason"`
	Severity  string                 `json:"severity"`
	UserVars  map[string]interface{} `json:"user-variables"`
	MessageID string                 `json:"message-id"` // Mailgun message ID
	Timestamp float64                `json:"timestamp"`  // Unix timestamp
}

func handleMailgunWebhook(c *gin.Context) {
	var payload MailgunWebhookPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	event := payload.EventData
	trackingService := service.NewWebhookTrackingService()

	// Parse user ID if provided
	var userID *uuid.UUID
	if userVarID, ok := event.UserVars["user_id"].(string); ok {
		if uid, err := uuid.Parse(userVarID); err == nil {
			userID = &uid
		}
	}

	// Create metadata
	metadata := map[string]interface{}{
		"event":    event.Event,
		"reason":   event.Reason,
		"severity": event.Severity,
		"provider": "mailgun",
	}

	// Process webhook event to update email status
	webhookEvent := service.EmailEvent{
		MessageID: event.MessageID,
		Email:     event.Recipient,
		Event:     event.Event,
		Reason:    event.Reason,
		Timestamp: time.Unix(int64(event.Timestamp), 0),
		Provider:  "mailgun",
		Metadata:  metadata,
		UserID:    userID,
	}

	if err := trackingService.ProcessWebhookEvent(webhookEvent); err != nil {
		// Log error but return success to provider
		c.JSON(http.StatusOK, gin.H{
			"status": "processed",
			"error":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "processed"})
}

// Generic webhook handler for custom integrations
func GenericWebhookHandler(c *gin.Context) {
	var payload struct {
		MessageID string                 `json:"message_id"`
		Email     string                 `json:"email" binding:"required"`
		Event     string                 `json:"event" binding:"required"`
		Reason    string                 `json:"reason"`
		UserID    string                 `json:"user_id"`
		Timestamp int64                  `json:"timestamp"`
		Metadata  map[string]interface{} `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	trackingService := service.NewWebhookTrackingService()

	// Parse user ID
	var userID *uuid.UUID
	if payload.UserID != "" {
		if uid, err := uuid.Parse(payload.UserID); err == nil {
			userID = &uid
		}
	}

	// Prepare metadata
	if payload.Metadata == nil {
		payload.Metadata = make(map[string]interface{})
	}
	payload.Metadata["event"] = payload.Event
	payload.Metadata["reason"] = payload.Reason
	payload.Metadata["provider"] = "generic"

	// Process webhook event to update email status
	webhookEvent := service.EmailEvent{
		MessageID: payload.MessageID,
		Email:     payload.Email,
		Event:     payload.Event,
		Reason:    payload.Reason,
		Timestamp: parseTimestamp(payload.Timestamp),
		Provider:  "generic",
		Metadata:  payload.Metadata,
		UserID:    userID,
	}

	if err := trackingService.ProcessWebhookEvent(webhookEvent); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "processed"})
}
