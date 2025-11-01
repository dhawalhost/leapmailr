package handlers

import (
	"net/http"

	"github.com/dhawalhost/leapmailr/models"
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

	suppressionService := service.NewSuppressionService()

	for _, event := range events {
		var reason models.SuppressionReason
		var shouldSuppress bool

		switch event.Event {
		case "bounce", "dropped":
			reason = models.SuppressionBounce
			shouldSuppress = true
		case "spamreport":
			reason = models.SuppressionComplaint
			shouldSuppress = true
		case "unsubscribe":
			reason = models.SuppressionUnsubscribe
			shouldSuppress = true
		}

		if shouldSuppress && event.Email != "" {
			metadata := map[string]interface{}{
				"event":     event.Event,
				"reason":    event.Reason,
				"status":    event.Status,
				"timestamp": event.Timestamp,
				"provider":  "sendgrid",
			}

			var userID *uuid.UUID
			if event.UserID != "" {
				if uid, err := uuid.Parse(event.UserID); err == nil {
					userID = &uid
				}
			}

			suppressionService.AddSuppressionFromWebhook(event.Email, reason, metadata, userID)
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": "processed"})
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
}

func handleMailgunWebhook(c *gin.Context) {
	var payload MailgunWebhookPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	event := payload.EventData
	suppressionService := service.NewSuppressionService()

	var reason models.SuppressionReason
	var shouldSuppress bool

	switch event.Event {
	case "failed", "bounced":
		if event.Severity == "permanent" {
			reason = models.SuppressionBounce
			shouldSuppress = true
		}
	case "complained":
		reason = models.SuppressionComplaint
		shouldSuppress = true
	case "unsubscribed":
		reason = models.SuppressionUnsubscribe
		shouldSuppress = true
	}

	if shouldSuppress && event.Recipient != "" {
		metadata := map[string]interface{}{
			"event":    event.Event,
			"reason":   event.Reason,
			"severity": event.Severity,
			"provider": "mailgun",
		}

		var userID *uuid.UUID
		if userVarID, ok := event.UserVars["user_id"].(string); ok {
			if uid, err := uuid.Parse(userVarID); err == nil {
				userID = &uid
			}
		}

		suppressionService.AddSuppressionFromWebhook(event.Recipient, reason, metadata, userID)
	}

	c.JSON(http.StatusOK, gin.H{"status": "processed"})
}

// Generic webhook handler for custom integrations
func GenericWebhookHandler(c *gin.Context) {
	var payload struct {
		Email    string                 `json:"email" binding:"required"`
		Event    string                 `json:"event" binding:"required"`
		Reason   string                 `json:"reason"`
		UserID   string                 `json:"user_id"`
		Metadata map[string]interface{} `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var reason models.SuppressionReason
	var shouldSuppress bool

	switch payload.Event {
	case "bounce":
		reason = models.SuppressionBounce
		shouldSuppress = true
	case "complaint", "spam":
		reason = models.SuppressionComplaint
		shouldSuppress = true
	case "unsubscribe":
		reason = models.SuppressionUnsubscribe
		shouldSuppress = true
	}

	if !shouldSuppress {
		c.JSON(http.StatusOK, gin.H{"status": "ignored"})
		return
	}

	suppressionService := service.NewSuppressionService()

	var userID *uuid.UUID
	if payload.UserID != "" {
		if uid, err := uuid.Parse(payload.UserID); err == nil {
			userID = &uid
		}
	}

	if payload.Metadata == nil {
		payload.Metadata = make(map[string]interface{})
	}
	payload.Metadata["event"] = payload.Event
	payload.Metadata["reason"] = payload.Reason

	err := suppressionService.AddSuppressionFromWebhook(payload.Email, reason, payload.Metadata, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "processed"})
}
