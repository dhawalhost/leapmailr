package service

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/dhawalhost/leapmailr/database"
	"github.com/dhawalhost/leapmailr/models"
	"github.com/google/uuid"
)

// CreateAutoReply creates a new auto-reply configuration
func CreateAutoReply(config *models.AutoReplyConfig) error {
	if config.UserID == uuid.Nil {
		return errors.New("user_id is required")
	}
	if config.Name == "" {
		return errors.New("name is required")
	}
	if config.Subject == "" {
		return errors.New("subject is required")
	}
	if config.Body == "" {
		return errors.New("body is required")
	}

	return database.DB.Create(config).Error
}

// GetAutoReply retrieves a single auto-reply configuration by ID
func GetAutoReply(id uuid.UUID, userID uuid.UUID) (*models.AutoReplyConfig, error) {
	var config models.AutoReplyConfig
	err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// ListAutoReplies lists all auto-reply configurations for a user
func ListAutoReplies(userID uuid.UUID) ([]models.AutoReplyConfig, error) {
	var configs []models.AutoReplyConfig
	err := database.DB.Where("user_id = ?", userID).Order("created_at DESC").Find(&configs).Error
	return configs, err
}

// UpdateAutoReply updates an existing auto-reply configuration
func UpdateAutoReply(id uuid.UUID, userID uuid.UUID, updates map[string]interface{}) error {
	result := database.DB.Model(&models.AutoReplyConfig{}).
		Where("id = ? AND user_id = ?", id, userID).
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("auto-reply not found")
	}
	return nil
}

// DeleteAutoReply deletes an auto-reply configuration
func DeleteAutoReply(id uuid.UUID, userID uuid.UUID) error {
	result := database.DB.Where("id = ? AND user_id = ?", id, userID).Delete(&models.AutoReplyConfig{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("auto-reply not found")
	}
	return nil
}

// GetActiveAutoReplyForService retrieves the active auto-reply for a specific email service
func GetActiveAutoReplyForService(userID uuid.UUID, serviceID uuid.UUID, isFormSubmission bool) (*models.AutoReplyConfig, error) {
	var config models.AutoReplyConfig
	query := database.DB.Where("user_id = ? AND is_active = ?", userID, true)

	// Check trigger type
	if isFormSubmission {
		query = query.Where("trigger_on_form = ?", true)
	} else {
		query = query.Where("trigger_on_api = ?", true)
	}

	// If serviceID is provided, prioritize service-specific auto-reply
	if serviceID != uuid.Nil {
		err := query.Where("email_service_id = ?", serviceID).First(&config).Error
		if err == nil {
			return &config, nil
		}
	}

	// Fall back to global auto-reply (no specific service)
	err := query.Where("email_service_id = ?", uuid.Nil).First(&config).Error
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// SendAutoReply sends an automatic reply to the specified recipient
func SendAutoReply(config *models.AutoReplyConfig, recipientEmail string, variables map[string]string) error {
	// Build the auto-reply email
	subject := config.Subject
	body := config.Body

	// Replace variables in subject and body if enabled
	if config.IncludeVariables && variables != nil {
		subject = replaceVariables(subject, variables)
		body = replaceVariables(body, variables)
	}

	// Apply delay if configured
	if config.DelaySeconds > 0 {
		time.Sleep(time.Duration(config.DelaySeconds) * time.Second)
	}

	// Create a temporary template for the auto-reply
	tempTemplate := models.Template{
		UserID:      &config.UserID,
		Name:        "Auto-Reply: " + config.Name,
		Subject:     subject,
		HTMLContent: body,
		TextContent: body,
	}

	// Create the template in database (we'll delete it after sending)
	if err := database.DB.Create(&tempTemplate).Error; err != nil {
		return fmt.Errorf("failed to create temporary template: %w", err)
	}
	defer database.DB.Delete(&tempTemplate) // Clean up temp template

	// Convert string map to interface map for TemplateParams
	templateParams := make(map[string]interface{})
	for k, v := range variables {
		templateParams[k] = v
	}

	// Create email request
	serviceID := config.EmailServiceID
	req := models.EmailRequest{
		TemplateID:     tempTemplate.ID,
		ServiceID:      &serviceID,
		ToEmail:        recipientEmail,
		Subject:        subject,
		TemplateParams: templateParams,
	}

	if config.FromEmail != "" {
		req.FromEmail = config.FromEmail
	}
	if config.FromName != "" {
		req.FromName = config.FromName
	}

	// Send the email using the existing email service
	emailSvc := NewEmailService()
	_, err := emailSvc.SendEmail(req, config.UserID)

	// Log the auto-reply
	logEntry := &models.AutoReplyLog{
		AutoReplyID:    config.ID,
		UserID:         config.UserID,
		RecipientEmail: recipientEmail,
		Subject:        subject,
		SentAt:         time.Now(),
	}

	if err != nil {
		logEntry.Status = "failed"
		logEntry.ErrorMessage = err.Error()
		database.DB.Create(logEntry)
		return err
	}

	logEntry.Status = "sent"
	database.DB.Create(logEntry)

	return nil
}

// replaceVariables replaces {{variable}} placeholders in the text with actual values
func replaceVariables(text string, variables map[string]string) string {
	if variables == nil {
		return text
	}

	// Pattern to match {{variable}}
	re := regexp.MustCompile(`\{\{([^}]+)\}\}`)

	result := re.ReplaceAllStringFunc(text, func(match string) string {
		// Extract variable name
		varName := strings.Trim(match, "{}")
		varName = strings.TrimSpace(varName)

		// Replace with value if exists, otherwise keep placeholder
		if value, exists := variables[varName]; exists {
			return value
		}
		return match
	})

	return result
}

// GetAutoReplyLogs retrieves auto-reply logs for a user
func GetAutoReplyLogs(userID uuid.UUID, limit int) ([]models.AutoReplyLog, error) {
	var logs []models.AutoReplyLog
	query := database.DB.Where("user_id = ?", userID).Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&logs).Error
	return logs, err
}

// GetAutoReplyLogsByConfig retrieves logs for a specific auto-reply configuration
func GetAutoReplyLogsByConfig(configID uuid.UUID, userID uuid.UUID, limit int) ([]models.AutoReplyLog, error) {
	var logs []models.AutoReplyLog
	query := database.DB.Where("autoreply_id = ? AND user_id = ?", configID, userID).Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&logs).Error
	return logs, err
}
