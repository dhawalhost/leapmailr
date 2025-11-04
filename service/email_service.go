package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/dhawalhost/leapmailr/database"
	"github.com/dhawalhost/leapmailr/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// EmailServiceManager handles email service operations
type EmailServiceManager struct {
	db *gorm.DB
}

// NewEmailServiceManager creates a new email service manager
func NewEmailServiceManager() *EmailServiceManager {
	return &EmailServiceManager{
		db: database.GetDB(),
	}
}

// CreateEmailService creates a new email service configuration
func (s *EmailServiceManager) CreateEmailService(req models.CreateEmailServiceRequest, userID uuid.UUID) (*models.EmailServiceResponse, error) {
	// Validate configuration based on provider
	if err := s.validateConfiguration(req.Provider, req.Configuration); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Marshal configuration to JSON
	configJSON, err := json.Marshal(req.Configuration)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal configuration: %w", err)
	}

	// If this is set as default, unset other defaults (within the same project if projectID is provided)
	if req.IsDefault {
		query := s.db.Model(&models.EmailService{}).
			Where("user_id = ? AND is_default = ?", userID, true)

		// If project_id is provided, only unset defaults within that project
		if req.ProjectID != nil {
			query = query.Where("project_id = ?", *req.ProjectID)
		} else {
			// If no project_id, unset defaults where project_id is NULL
			query = query.Where("project_id IS NULL")
		}

		if err := query.Update("is_default", false).Error; err != nil {
			return nil, fmt.Errorf("failed to unset previous default: %w", err)
		}
	}

	emailService := models.EmailService{
		UserID:        &userID,
		ProjectID:     req.ProjectID,
		Name:          req.Name,
		Provider:      req.Provider,
		Configuration: string(configJSON),
		FromEmail:     req.FromEmail,
		FromName:      req.FromName,
		ReplyToEmail:  req.ReplyToEmail,
		IsDefault:     req.IsDefault,
		Status:        "active",
	}

	if err := s.db.Create(&emailService).Error; err != nil {
		return nil, fmt.Errorf("failed to create email service: %w", err)
	}

	return s.toResponse(&emailService), nil
}

// GetEmailService retrieves an email service by ID
func (s *EmailServiceManager) GetEmailService(serviceID, userID uuid.UUID) (*models.EmailServiceResponse, error) {
	var service models.EmailService
	err := s.db.Where("id = ? AND user_id = ?", serviceID, userID).First(&service).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("email service not found")
		}
		return nil, fmt.Errorf("failed to get email service: %w", err)
	}

	return s.toResponse(&service), nil
}

// ListEmailServices retrieves all email services for a user
func (s *EmailServiceManager) ListEmailServices(userID uuid.UUID, filters models.EmailServiceFilters) ([]models.EmailServiceResponse, error) {
	var services []models.EmailService
	query := s.db.Where("user_id = ?", userID)

	if filters.ProjectID != nil {
		query = query.Where("project_id = ?", *filters.ProjectID)
	}

	if filters.Provider != "" {
		query = query.Where("provider = ?", filters.Provider)
	}

	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}

	query = query.Order("is_default DESC, created_at DESC")

	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	}

	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}

	if err := query.Find(&services).Error; err != nil {
		return nil, fmt.Errorf("failed to list email services: %w", err)
	}

	responses := make([]models.EmailServiceResponse, len(services))
	for i, service := range services {
		responses[i] = *s.toResponse(&service)
	}

	return responses, nil
}

// UpdateEmailService updates an existing email service
func (s *EmailServiceManager) UpdateEmailService(serviceID uuid.UUID, req models.UpdateEmailServiceRequest, userID uuid.UUID) (*models.EmailServiceResponse, error) {
	var service models.EmailService
	err := s.db.Where("id = ? AND user_id = ?", serviceID, userID).First(&service).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("email service not found")
		}
		return nil, fmt.Errorf("failed to get email service: %w", err)
	}

	// Update fields
	if req.Name != "" {
		service.Name = req.Name
	}

	if req.FromEmail != "" {
		service.FromEmail = req.FromEmail
	}

	if req.FromName != "" {
		service.FromName = req.FromName
	}

	if req.ReplyToEmail != "" {
		service.ReplyToEmail = req.ReplyToEmail
	}

	if req.Configuration != nil {
		// Validate configuration
		if err := s.validateConfiguration(service.Provider, req.Configuration); err != nil {
			return nil, fmt.Errorf("invalid configuration: %w", err)
		}

		configJSON, err := json.Marshal(req.Configuration)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal configuration: %w", err)
		}
		service.Configuration = string(configJSON)
	}

	if req.IsDefault != nil && *req.IsDefault {
		// Unset other defaults
		if err := s.db.Model(&models.EmailService{}).
			Where("user_id = ? AND id != ? AND is_default = ?", userID, serviceID, true).
			Update("is_default", false).Error; err != nil {
			return nil, fmt.Errorf("failed to unset previous default: %w", err)
		}
		service.IsDefault = true
	} else if req.IsDefault != nil {
		service.IsDefault = false
	}

	if req.Status != "" {
		service.Status = req.Status
	}

	service.UpdatedAt = time.Now()

	if err := s.db.Save(&service).Error; err != nil {
		return nil, fmt.Errorf("failed to update email service: %w", err)
	}

	return s.toResponse(&service), nil
}

// DeleteEmailService deletes an email service
func (s *EmailServiceManager) DeleteEmailService(serviceID, userID uuid.UUID) error {
	result := s.db.Where("id = ? AND user_id = ?", serviceID, userID).Delete(&models.EmailService{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete email service: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("email service not found")
	}
	return nil
}

// TestEmailService tests an email service by sending a test email
func (s *EmailServiceManager) TestEmailService(serviceID, userID uuid.UUID, req models.TestEmailServiceRequest) (*models.TestEmailServiceResponse, error) {
	var service models.EmailService
	err := s.db.Where("id = ? AND user_id = ?", serviceID, userID).First(&service).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &models.TestEmailServiceResponse{
				Success: false,
				Error:   "Email service not found",
			}, nil
		}
		return nil, fmt.Errorf("failed to get email service: %w", err)
	}

	// Parse configuration
	var config map[string]interface{}
	if err := json.Unmarshal([]byte(service.Configuration), &config); err != nil {
		return &models.TestEmailServiceResponse{
			Success: false,
			Error:   "Failed to parse service configuration",
		}, nil
	}

	// Validate configuration
	if err := s.validateConfiguration(service.Provider, config); err != nil {
		// Update service with error
		s.db.Model(&service).Updates(map[string]interface{}{
			"status":     "error",
			"last_error": err.Error(),
		})

		return &models.TestEmailServiceResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// Actually send a test email
	emailService := NewEmailService()

	// Create a simple test email log
	fromEmail := ""
	if fe, ok := config["from_email"].(string); ok {
		fromEmail = fe
	}
	if fromEmail == "" {
		fromEmail = "noreply@example.com"
	}

	emailLog := models.EmailLog{
		UserID:    &userID,
		ServiceID: &service.ID,
		FromEmail: fromEmail,
		FromName:  "LeapMailr Test",
		ToEmail:   req.ToEmail,
		ToName:    "Test Recipient",
		Subject:   "Test Email from LeapMailr",
		Status:    "queued",
		Metadata:  "{}", // Initialize with empty JSON object
	}

	if err := s.db.Create(&emailLog).Error; err != nil {
		return &models.TestEmailServiceResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to create email log: %v", err),
		}, nil
	}

	// Send test email
	htmlContent := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 30px; text-align: center; border-radius: 8px 8px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border-radius: 0 0 8px 8px; }
        .success { background: #d4edda; border: 1px solid #c3e6cb; color: #155724; padding: 15px; border-radius: 4px; margin: 20px 0; }
        .info { background: #fff; border: 1px solid #ddd; padding: 15px; border-radius: 4px; margin: 15px 0; }
        .footer { text-align: center; color: #666; font-size: 12px; margin-top: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>✅ Email Service Test</h1>
        </div>
        <div class="content">
            <div class="success">
                <strong>Success!</strong> Your email service is working correctly.
            </div>
            <p>This is a test email from your LeapMailr email service configuration.</p>
            <div class="info">
                <strong>Service Name:</strong> ` + service.Name + `<br>
                <strong>Provider:</strong> ` + service.Provider + `<br>
                <strong>Status:</strong> Active
            </div>
            <p>If you received this email, your SMTP configuration is set up correctly and you can start sending emails through LeapMailr!</p>
        </div>
        <div class="footer">
            <p>This is an automated test email from LeapMailr</p>
        </div>
    </div>
</body>
</html>`

	textContent := fmt.Sprintf(`EMAIL SERVICE TEST

Success! Your email service is working correctly.

This is a test email from your LeapMailr email service configuration.

Service Name: %s
Provider: %s
Status: Active

If you received this email, your SMTP configuration is set up correctly and you can start sending emails through LeapMailr!

---
This is an automated test email from LeapMailr`, service.Name, service.Provider)

	// Create a dummy template for test email (test emails don't use actual templates)
	dummyTemplate := models.Template{
		FromEmail:    service.FromEmail,
		FromName:     service.FromName,
		ReplyToEmail: service.ReplyToEmail,
	}

	// Send the email
	err = emailService.sendEmailViaSMTP(service, dummyTemplate, emailLog, htmlContent, textContent, nil)
	if err != nil {
		emailLog.Status = "failed"
		emailLog.ErrorMessage = err.Error()
		s.db.Save(&emailLog)

		// Update service with error
		s.db.Model(&service).Updates(map[string]interface{}{
			"status":     "error",
			"last_error": err.Error(),
		})

		return &models.TestEmailServiceResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to send test email: %v", err),
		}, nil
	}

	// Update email log as sent
	emailLog.Status = "sent"
	now := time.Now()
	emailLog.SentAt = &now
	s.db.Save(&emailLog)

	// Update service status
	s.db.Model(&service).Updates(map[string]interface{}{
		"status":     "active",
		"last_error": "",
	})

	return &models.TestEmailServiceResponse{
		Success: true,
		Message: fmt.Sprintf("Test email sent successfully to %s", req.ToEmail),
	}, nil
}

// SetDefault sets an email service as the default
func (s *EmailServiceManager) SetDefault(serviceID, userID uuid.UUID) error {
	// First, verify the service exists and belongs to the user
	var service models.EmailService
	err := s.db.Where("id = ? AND user_id = ?", serviceID, userID).First(&service).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("email service not found")
		}
		return fmt.Errorf("failed to get email service: %w", err)
	}

	// Unset all other defaults for this user
	if err := s.db.Model(&models.EmailService{}).
		Where("user_id = ? AND id != ?", userID, serviceID).
		Update("is_default", false).Error; err != nil {
		return fmt.Errorf("failed to unset previous defaults: %w", err)
	}

	// Set this one as default
	if err := s.db.Model(&service).Update("is_default", true).Error; err != nil {
		return fmt.Errorf("failed to set default: %w", err)
	}

	return nil
}

// GetDefaultService gets the default email service for a user
func (s *EmailServiceManager) GetDefaultService(userID uuid.UUID) (*models.EmailService, error) {
	var service models.EmailService
	err := s.db.Where("user_id = ? AND is_default = ?", userID, true).First(&service).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// If no default, get the first active service
			err = s.db.Where("user_id = ? AND status = ?", userID, "active").
				Order("created_at ASC").
				First(&service).Error
			if err != nil {
				return nil, fmt.Errorf("no email service configured")
			}
		} else {
			return nil, fmt.Errorf("failed to get default service: %w", err)
		}
	}

	return &service, nil
}

// Helper functions

func (s *EmailServiceManager) toResponse(service *models.EmailService) *models.EmailServiceResponse {
	response := &models.EmailServiceResponse{
		ID:           service.ID,
		Name:         service.Name,
		Provider:     service.Provider,
		FromEmail:    service.FromEmail,
		FromName:     service.FromName,
		ReplyToEmail: service.ReplyToEmail,
		IsDefault:    service.IsDefault,
		Status:       service.Status,
		LastError:    service.LastError,
		CreatedAt:    service.CreatedAt,
		UpdatedAt:    service.UpdatedAt,
	}

	// Add provider metadata (logo and color)
	metadata := getProviderMetadata(service.Provider)
	response.ProviderLogo = metadata.Logo
	response.ProviderColor = metadata.Color

	// Add safe config preview
	var config map[string]interface{}
	if err := json.Unmarshal([]byte(service.Configuration), &config); err == nil {
		response.ConfigPreview = s.createConfigPreview(service.Provider, config)
	}

	return response
}

// ProviderMetadata holds display metadata for each provider
type ProviderMetadata struct {
	Logo  string
	Color string
}

// getProviderMetadata returns the logo SVG and brand color for each provider
func getProviderMetadata(provider string) ProviderMetadata {
	metadata := map[string]ProviderMetadata{
		"smtp": {
			Logo:  `<svg viewBox="0 0 24 24" fill="none"><rect x="3" y="5" width="18" height="14" rx="2" stroke="currentColor" stroke-width="2"/><path d="M3 8l9 6 9-6" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg>`,
			Color: "#4A90E2",
		},
		"sendgrid": {
			Logo:  `<svg viewBox="0 0 200 200" fill="currentColor"><path d="M50 50h33.33v33.33H50V50zm0 66.67h33.33V150H50v-33.33zm66.67-66.67H150v33.33h-33.33V50zm0 66.67H150V150h-33.33v-33.33zM50 116.67h33.33v-33.34H50v33.34zm66.67 0H150v-33.34h-33.33v33.34z"/></svg>`,
			Color: "#1A82E2",
		},
		"mailgun": {
			Logo:  `<svg viewBox="0 0 200 200" fill="currentColor"><circle cx="100" cy="100" r="80" fill="none" stroke="currentColor" stroke-width="8"/><circle cx="100" cy="100" r="35"/><path d="M100 20v20m0 120v20m80-80h-20M20 100h20" stroke="currentColor" stroke-width="8" stroke-linecap="round"/></svg>`,
			Color: "#F06B66",
		},
		"ses": {
			Logo:  `<svg viewBox="0 0 200 200" fill="currentColor"><path d="M100 20L40 60v50l60 40 60-40V60L100 20z"/><path d="M100 80L60 100v40l40 25 40-25v-40L100 80z" opacity="0.6"/><path d="M55 115l-10 7 4 12 10-7-4-12m90-25l10-7-4-12-10 7 4 12M25 90l12 3 2-12-12-3-2 12m150 20l-12-3-2 12 12 3 2-12"/></svg>`,
			Color: "#FF9900",
		},
		"postmark": {
			Logo:  `<svg viewBox="0 0 200 200" fill="currentColor"><path d="M160 40H40c-8 0-16 8-16 16v88c0 8 8 16 16 16h120c8 0 16-8 16-16V56c0-8-8-16-16-16zm0 32l-60 40-60-40V56l60 40 60-40v16z"/><circle cx="100" cy="100" r="15" fill="white"/></svg>`,
			Color: "#FFCC00",
		},
		"resend": {
			Logo:  `<svg viewBox="0 0 200 200" fill="currentColor"><path d="M20 160l160-80-160-80v60l120 20-120 20v60z"/><path d="M140 100l-20 20v-40l20 20z" opacity="0.7"/></svg>`,
			Color: "#000000",
		},
	}

	if meta, ok := metadata[provider]; ok {
		return meta
	}

	// Default fallback
	return ProviderMetadata{
		Logo:  `<svg viewBox="0 0 24 24" fill="none"><rect x="3" y="5" width="18" height="14" rx="2" stroke="currentColor" stroke-width="2"/><path d="M3 8l9 6 9-6" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg>`,
		Color: "#4A90E2",
	}
}

func (s *EmailServiceManager) createConfigPreview(provider string, config map[string]interface{}) map[string]string {
	preview := make(map[string]string)

	switch provider {
	case "smtp":
		if host, ok := config["host"].(string); ok {
			preview["host"] = host
		}
		if port, ok := config["port"].(float64); ok {
			preview["port"] = fmt.Sprintf("%.0f", port)
		}
		if username, ok := config["username"].(string); ok {
			preview["username"] = maskString(username)
		}
		if fromEmail, ok := config["from_email"].(string); ok {
			preview["from_email"] = fromEmail
		}

	case "sendgrid":
		if fromEmail, ok := config["from_email"].(string); ok {
			preview["from_email"] = fromEmail
		}
		preview["api_key"] = "***"

	case "mailgun":
		if domain, ok := config["domain"].(string); ok {
			preview["domain"] = domain
		}
		if fromEmail, ok := config["from_email"].(string); ok {
			preview["from_email"] = fromEmail
		}
		preview["api_key"] = "***"

	case "ses":
		if region, ok := config["region"].(string); ok {
			preview["region"] = region
		}
		if fromEmail, ok := config["from_email"].(string); ok {
			preview["from_email"] = fromEmail
		}
		preview["access_key"] = "***"

	case "postmark", "resend":
		if fromEmail, ok := config["from_email"].(string); ok {
			preview["from_email"] = fromEmail
		}
		preview["api_key"] = "***"
	}

	return preview
}

func (s *EmailServiceManager) validateConfiguration(provider string, config map[string]interface{}) error {
	switch provider {
	case "smtp":
		// from_email is now a separate field on EmailService, not in configuration
		required := []string{"host", "port", "username", "password"}
		for _, field := range required {
			if _, ok := config[field]; !ok {
				return fmt.Errorf("missing required field: %s", field)
			}
		}

	case "sendgrid":
		// from_email is now a separate field on EmailService, not in configuration
		required := []string{"api_key"}
		for _, field := range required {
			if _, ok := config[field]; !ok {
				return fmt.Errorf("missing required field: %s", field)
			}
		}

	case "mailgun":
		// from_email is now a separate field on EmailService, not in configuration
		required := []string{"domain", "api_key"}
		for _, field := range required {
			if _, ok := config[field]; !ok {
				return fmt.Errorf("missing required field: %s", field)
			}
		}

	case "ses":
		// from_email is now a separate field on EmailService, not in configuration
		required := []string{"region", "access_key", "secret_key"}
		for _, field := range required {
			if _, ok := config[field]; !ok {
				return fmt.Errorf("missing required field: %s", field)
			}
		}

	case "postmark":
		// from_email is now a separate field on EmailService, not in configuration
		required := []string{"server_token"}
		for _, field := range required {
			if _, ok := config[field]; !ok {
				return fmt.Errorf("missing required field: %s", field)
			}
		}

	case "resend":
		// from_email is now a separate field on EmailService, not in configuration
		required := []string{"api_key"}
		for _, field := range required {
			if _, ok := config[field]; !ok {
				return fmt.Errorf("missing required field: %s", field)
			}
		}

	default:
		return fmt.Errorf("unsupported provider: %s", provider)
	}

	return nil
}

func maskString(s string) string {
	if len(s) <= 4 {
		return "***"
	}
	return s[:2] + "***" + s[len(s)-2:]
}
