package service

import (
	"fmt"
	"time"

	"github.com/dhawalhost/leapmailr/database"
	"github.com/dhawalhost/leapmailr/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// WebhookTrackingService handles email delivery tracking via webhooks
type WebhookTrackingService struct {
	db *gorm.DB
}

// NewWebhookTrackingService creates a new webhook tracking service
func NewWebhookTrackingService() *WebhookTrackingService {
	return &WebhookTrackingService{
		db: database.GetDB(),
	}
}

// EmailEvent represents a normalized email event from any provider
type EmailEvent struct {
	MessageID string                 // Provider's message ID
	Email     string                 // Recipient email
	Event     string                 // Event type: delivered, bounced, opened, clicked, failed, complained, unsubscribed
	Status    string                 // Email status to set
	Reason    string                 // Failure/bounce reason
	Timestamp time.Time              // Event timestamp
	Provider  string                 // Provider name (sendgrid, mailgun, smtp, etc.)
	Metadata  map[string]interface{} // Additional event data
	UserID    *uuid.UUID             // Optional user ID
}

// ProcessWebhookEvent processes a webhook event and updates email status
func (s *WebhookTrackingService) ProcessWebhookEvent(event EmailEvent) error {
	// Find email log by message ID or recipient email
	var emailLog models.EmailLog
	var err error

	if event.MessageID != "" {
		// Try to find by message ID first (most reliable)
		err = s.db.Where("message_id = ?", event.MessageID).First(&emailLog).Error
	} else if event.Email != "" {
		// Fallback to finding by email (less reliable, get most recent)
		err = s.db.Where("to_email = ?", event.Email).Order("created_at DESC").First(&emailLog).Error
	} else {
		return fmt.Errorf("webhook event missing both message_id and email")
	}

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// This might be an email sent outside our system, ignore gracefully
			return nil
		}
		return fmt.Errorf("failed to find email log: %w", err)
	}

	// Update email status based on event type
	now := time.Now()
	updates := make(map[string]interface{})

	switch event.Event {
	case "delivered":
		updates["status"] = "delivered"
		updates["delivered_at"] = now
		if emailLog.SentAt == nil {
			updates["sent_at"] = now
		}

	case "bounced", "bounce", "failed", "dropped":
		updates["status"] = "bounced"
		if event.Reason != "" {
			updates["error_message"] = event.Reason
		}

	case "opened", "open":
		// Only update if not already in a final state
		if emailLog.Status != "bounced" && emailLog.Status != "failed" {
			updates["status"] = "opened"
			updates["opened_at"] = now
			if emailLog.DeliveredAt == nil {
				updates["delivered_at"] = now
			}
		}

	case "clicked", "click":
		// Only update if not already in a final state
		if emailLog.Status != "bounced" && emailLog.Status != "failed" {
			updates["status"] = "clicked"
			updates["clicked_at"] = now
			if emailLog.OpenedAt == nil {
				updates["opened_at"] = now
			}
			if emailLog.DeliveredAt == nil {
				updates["delivered_at"] = now
			}
		}

	case "complained", "spamreport", "spam":
		updates["status"] = "failed"
		updates["error_message"] = "Marked as spam"
		// Also add to suppression list
		if emailLog.UserID != nil {
			suppressionService := NewSuppressionService()
			suppressionService.AddSuppressionFromWebhook(
				emailLog.ToEmail,
				models.SuppressionComplaint,
				event.Metadata,
				emailLog.UserID,
			)
		}

	case "unsubscribed", "unsubscribe":
		// Add to suppression list
		if emailLog.UserID != nil {
			suppressionService := NewSuppressionService()
			suppressionService.AddSuppressionFromWebhook(
				emailLog.ToEmail,
				models.SuppressionUnsubscribe,
				event.Metadata,
				emailLog.UserID,
			)
		}

	default:
		// Unknown event type, log but don't fail
		return nil
	}

	// Update the email log
	if len(updates) > 0 {
		if err := s.db.Model(&emailLog).Updates(updates).Error; err != nil {
			return fmt.Errorf("failed to update email log: %w", err)
		}
	}

	return nil
}

// UpdateEmailStatusByMessageID updates email status by message ID
func (s *WebhookTrackingService) UpdateEmailStatusByMessageID(messageID, status string, timestamp time.Time) error {
	var emailLog models.EmailLog
	if err := s.db.Where("message_id = ?", messageID).First(&emailLog).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil // Email not found in our system, ignore
		}
		return fmt.Errorf("failed to find email log: %w", err)
	}

	updates := map[string]interface{}{
		"status": status,
	}

	// Set appropriate timestamp based on status
	switch status {
	case "delivered":
		updates["delivered_at"] = timestamp
	case "opened":
		updates["opened_at"] = timestamp
	case "clicked":
		updates["clicked_at"] = timestamp
	}

	if err := s.db.Model(&emailLog).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update email status: %w", err)
	}

	return nil
}

// UpdateEmailStatusByEmail updates email status by recipient email (fallback method)
func (s *WebhookTrackingService) UpdateEmailStatusByEmail(email, status string, timestamp time.Time) error {
	// Get the most recent email to this address
	var emailLog models.EmailLog
	if err := s.db.Where("to_email = ?", email).Order("created_at DESC").First(&emailLog).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil // Email not found, ignore
		}
		return fmt.Errorf("failed to find email log: %w", err)
	}

	updates := map[string]interface{}{
		"status": status,
	}

	switch status {
	case "delivered":
		updates["delivered_at"] = timestamp
	case "opened":
		updates["opened_at"] = timestamp
	case "clicked":
		updates["clicked_at"] = timestamp
	}

	if err := s.db.Model(&emailLog).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update email status: %w", err)
	}

	return nil
}

// GetEmailStats returns email statistics for a user
func (s *WebhookTrackingService) GetEmailStats(userID uuid.UUID, startDate, endDate time.Time) (map[string]int64, error) {
	stats := make(map[string]int64)

	query := s.db.Model(&models.EmailLog{}).Where("user_id = ?", userID)
	if !startDate.IsZero() {
		query = query.Where("created_at >= ?", startDate)
	}
	if !endDate.IsZero() {
		query = query.Where("created_at <= ?", endDate)
	}

	// Count by status
	var results []struct {
		Status string
		Count  int64
	}

	if err := query.Select("status, COUNT(*) as count").Group("status").Find(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to get email stats: %w", err)
	}

	for _, r := range results {
		stats[r.Status] = r.Count
	}

	return stats, nil
}
