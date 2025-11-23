package service

import (
	"fmt"
	"time"

	"github.com/dhawalhost/leapmailr/database"
	"github.com/dhawalhost/leapmailr/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Webhook error messages
const (
	errFailedFindEmailLog = "failed to find email log: %w"
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
	emailLog, err := s.findEmailLog(event.MessageID, event.Email)
	if err != nil {
		return err
	}
	if emailLog == nil {
		// Email not found in our system, ignore gracefully
		return nil
	}

	// Build updates based on event type
	updates := s.buildUpdatesForEvent(event, emailLog)

	// Handle suppression list updates for spam/unsubscribe
	s.handleSuppressionEvents(event, emailLog)

	// Update the email log
	if len(updates) > 0 {
		if err := s.db.Model(emailLog).Updates(updates).Error; err != nil {
			return fmt.Errorf("failed to update email log: %w", err)
		}
	}

	return nil
}

// findEmailLog finds an email log by message ID or email address
func (s *WebhookTrackingService) findEmailLog(messageID, email string) (*models.EmailLog, error) {
	var emailLog models.EmailLog
	var err error

	if messageID != "" {
		// Try to find by message ID first (most reliable)
		err = s.db.Where("message_id = ?", messageID).First(&emailLog).Error
	} else if email != "" {
		// Fallback to finding by email (less reliable, get most recent)
		err = s.db.Where("to_email = ?", email).Order("created_at DESC").First(&emailLog).Error
	} else {
		return nil, fmt.Errorf("webhook event missing both message_id and email")
	}

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf(errFailedFindEmailLog, err)
	}

	return &emailLog, nil
}

// buildUpdatesForEvent creates update map based on event type
func (s *WebhookTrackingService) buildUpdatesForEvent(event EmailEvent, emailLog *models.EmailLog) map[string]interface{} {
	now := time.Now()
	updates := make(map[string]interface{})

	switch event.Event {
	case "delivered":
		s.handleDeliveredEvent(updates, emailLog, now)
	case "bounced", "bounce", "failed", "dropped":
		s.handleBouncedEvent(updates, event)
	case "opened", "open":
		s.handleOpenedEvent(updates, emailLog, now)
	case "clicked", "click":
		s.handleClickedEvent(updates, emailLog, now)
	case "complained", "spamreport", "spam":
		s.handleComplaintEvent(updates)
	case "unsubscribed", "unsubscribe":
		// Handled in handleSuppressionEvents
	default:
		// Unknown event type, log but don't fail
	}

	return updates
}

// handleDeliveredEvent handles delivered event updates
func (s *WebhookTrackingService) handleDeliveredEvent(updates map[string]interface{}, emailLog *models.EmailLog, now time.Time) {
	updates["status"] = "delivered"
	updates["delivered_at"] = now
	if emailLog.SentAt == nil {
		updates["sent_at"] = now
	}
}

// handleBouncedEvent handles bounced/failed event updates
func (s *WebhookTrackingService) handleBouncedEvent(updates map[string]interface{}, event EmailEvent) {
	updates["status"] = "bounced"
	if event.Reason != "" {
		updates["error_message"] = event.Reason
	}
}

// handleOpenedEvent handles opened event updates
func (s *WebhookTrackingService) handleOpenedEvent(updates map[string]interface{}, emailLog *models.EmailLog, now time.Time) {
	// Only update if not already in a final state
	if emailLog.Status != "bounced" && emailLog.Status != "failed" {
		updates["status"] = "opened"
		updates["opened_at"] = now
		if emailLog.DeliveredAt == nil {
			updates["delivered_at"] = now
		}
	}
}

// handleClickedEvent handles clicked event updates
func (s *WebhookTrackingService) handleClickedEvent(updates map[string]interface{}, emailLog *models.EmailLog, now time.Time) {
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
}

// handleComplaintEvent handles spam complaint event updates
func (s *WebhookTrackingService) handleComplaintEvent(updates map[string]interface{}) {
	updates["status"] = "failed"
	updates["error_message"] = "Marked as spam"
}

// handleSuppressionEvents adds email to suppression list for spam/unsubscribe events
func (s *WebhookTrackingService) handleSuppressionEvents(event EmailEvent, emailLog *models.EmailLog) {
	if emailLog.UserID == nil {
		return
	}

	suppressionService := NewSuppressionService()

	switch event.Event {
	case "complained", "spamreport", "spam":
		suppressionService.AddSuppressionFromWebhook(
			emailLog.ToEmail,
			models.SuppressionComplaint,
			event.Metadata,
			emailLog.UserID,
		)
	case "unsubscribed", "unsubscribe":
		suppressionService.AddSuppressionFromWebhook(
			emailLog.ToEmail,
			models.SuppressionUnsubscribe,
			event.Metadata,
			emailLog.UserID,
		)
	}
}

// UpdateEmailStatusByMessageID updates email status by message ID
func (s *WebhookTrackingService) UpdateEmailStatusByMessageID(messageID, status string, timestamp time.Time) error {
	var emailLog models.EmailLog
	if err := s.db.Where("message_id = ?", messageID).First(&emailLog).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil // Email not found in our system, ignore
		}
		return fmt.Errorf(errFailedFindEmailLog, err)
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
		return fmt.Errorf(errFailedFindEmailLog, err)
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
