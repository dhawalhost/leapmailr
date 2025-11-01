package service

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/smtp"
	"strings"
	"text/template"
	"time"

	"github.com/dhawalhost/leapmailr/database"
	"github.com/dhawalhost/leapmailr/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// EmailService handles email operations
type EmailService struct {
	db *gorm.DB
}

// NewEmailService creates a new email service
func NewEmailService() *EmailService {
	return &EmailService{
		db: database.GetDB(),
	}
}

// SendEmail sends an email using the new API
func (s *EmailService) SendEmail(req models.EmailRequest, userID uuid.UUID) (*models.EmailLog, error) {
	// Check suppression list first
	suppressionService := NewSuppressionService()
	isSuppressed, suppression, err := suppressionService.IsEmailSuppressed(req.ToEmail, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check suppression list: %w", err)
	}
	if isSuppressed {
		return nil, fmt.Errorf("email address is suppressed (reason: %s)", suppression.Reason)
	}

	// Get template
	var emailTemplate models.Template
	if err := s.db.Where("id = ? AND (user_id = ? OR organization_id IN (SELECT organization_id FROM user_organizations WHERE user_id = ?))",
		req.TemplateID, userID, userID).First(&emailTemplate).Error; err != nil {
		return nil, fmt.Errorf("template not found: %w", err)
	}

	// Get email service (use default if not specified)
	var emailService models.EmailService
	if req.ServiceID != nil {
		if err := s.db.Where("id = ? AND (user_id = ? OR organization_id IN (SELECT organization_id FROM user_organizations WHERE user_id = ?))",
			*req.ServiceID, userID, userID).First(&emailService).Error; err != nil {
			return nil, fmt.Errorf("email service not found: %w", err)
		}
	} else {
		// Use default service
		if err := s.db.Where("(user_id = ? OR organization_id IN (SELECT organization_id FROM user_organizations WHERE user_id = ?)) AND is_default = ?",
			userID, userID, true).First(&emailService).Error; err != nil {
			return nil, fmt.Errorf("no default email service found: %w", err)
		}
	}

	// Process template
	subject := req.Subject
	if subject == "" {
		subject = emailTemplate.Subject
	}

	htmlContent, textContent, err := s.processTemplate(emailTemplate, req.TemplateParams, subject)
	if err != nil {
		return nil, fmt.Errorf("failed to process template: %w", err)
	}

	// Create email log entry
	emailLog := models.EmailLog{
		UserID:     &userID,
		TemplateID: &emailTemplate.ID,
		ServiceID:  &emailService.ID,
		FromEmail:  getFromEmail(req.FromEmail, emailService),
		FromName:   getFromName(req.FromName, emailService),
		ToEmail:    req.ToEmail,
		ToName:     req.ToName,
		Subject:    subject,
		Status:     "queued",
		Metadata:   "{}",
	}

	if err := s.db.Create(&emailLog).Error; err != nil {
		return nil, fmt.Errorf("failed to create email log: %w", err)
	}

	// Send email immediately or schedule
	if req.ScheduleAt != nil && req.ScheduleAt.After(time.Now()) {
		emailLog.Status = "scheduled"
		s.db.Save(&emailLog)
		// TODO: Add to job queue for scheduled sending
		return &emailLog, nil
	}

	// Send email now
	log.Printf("Attempting to send email to %s via service %s", emailLog.ToEmail, emailService.Name)
	err = s.sendEmailViaSMTP(emailService, emailLog, htmlContent, textContent, req.Attachments)
	if err != nil {
		log.Printf("Failed to send email: %v", err)
		emailLog.Status = "failed"
		emailLog.ErrorMessage = err.Error()
		s.db.Save(&emailLog)
		return &emailLog, err
	}

	log.Printf("Email sent successfully to %s", emailLog.ToEmail)
	emailLog.Status = "sent"
	emailLog.SentAt = &time.Time{}
	*emailLog.SentAt = time.Now()
	s.db.Save(&emailLog)

	return &emailLog, nil
}

// SendBulkEmail sends bulk emails
func (s *EmailService) SendBulkEmail(req models.BulkEmailRequest, userID uuid.UUID) ([]models.EmailLog, error) {
	var emailLogs []models.EmailLog

	// Get template
	var emailTemplate models.Template
	if err := s.db.Where("id = ? AND (user_id = ? OR organization_id IN (SELECT organization_id FROM user_organizations WHERE user_id = ?))",
		req.TemplateID, userID, userID).First(&emailTemplate).Error; err != nil {
		return nil, fmt.Errorf("template not found: %w", err)
	}

	// Get email service
	var emailService models.EmailService
	if req.ServiceID != nil {
		if err := s.db.Where("id = ? AND (user_id = ? OR organization_id IN (SELECT organization_id FROM user_organizations WHERE user_id = ?))",
			*req.ServiceID, userID, userID).First(&emailService).Error; err != nil {
			return nil, fmt.Errorf("email service not found: %w", err)
		}
	} else {
		if err := s.db.Where("(user_id = ? OR organization_id IN (SELECT organization_id FROM user_organizations WHERE user_id = ?)) AND is_default = ?",
			userID, userID, true).First(&emailService).Error; err != nil {
			return nil, fmt.Errorf("no default email service found: %w", err)
		}
	}

	// Process each recipient
	suppressionService := NewSuppressionService()
	for _, recipient := range req.Recipients {
		// Check suppression list
		isSuppressed, _, err := suppressionService.IsEmailSuppressed(recipient.Email, userID)
		if err != nil || isSuppressed {
			// Skip suppressed emails
			continue
		}

		// Merge default params with recipient-specific params
		templateParams := make(map[string]interface{})
		for k, v := range req.DefaultParams {
			templateParams[k] = v
		}
		for k, v := range recipient.TemplateParams {
			templateParams[k] = v
		}

		subject := req.Subject
		if subject == "" {
			subject = emailTemplate.Subject
		}

		htmlContent, textContent, err := s.processTemplate(emailTemplate, templateParams, subject)
		if err != nil {
			continue // Skip this recipient on template error
		}

		// Create email log entry
		emailLog := models.EmailLog{
			UserID:     &userID,
			TemplateID: &emailTemplate.ID,
			ServiceID:  &emailService.ID,
			FromEmail:  getFromEmail(req.FromEmail, emailService),
			FromName:   getFromName(req.FromName, emailService),
			ToEmail:    recipient.Email,
			ToName:     recipient.Name,
			Subject:    subject,
			Status:     "queued",
			Metadata:   "{}",
		}

		if err := s.db.Create(&emailLog).Error; err != nil {
			continue // Skip on database error
		}

		emailLogs = append(emailLogs, emailLog)

		// Schedule or send immediately
		if req.ScheduleAt != nil && req.ScheduleAt.After(time.Now()) {
			emailLog.Status = "scheduled"
			s.db.Save(&emailLog)
			// TODO: Add to job queue
			continue
		}

		// Send email (in production, this should be queued)
		go func(log models.EmailLog, html, text string) {
			err := s.sendEmailViaSMTP(emailService, log, html, text, nil)
			if err != nil {
				log.Status = "failed"
				log.ErrorMessage = err.Error()
			} else {
				log.Status = "sent"
				now := time.Now()
				log.SentAt = &now
			}
			s.db.Save(&log)
		}(emailLog, htmlContent, textContent)
	}

	return emailLogs, nil
}

// processTemplate processes a template with given parameters
func (s *EmailService) processTemplate(tmpl models.Template, params map[string]interface{}, subject string) (string, string, error) {
	// Process HTML content
	var htmlContent string
	if tmpl.HTMLContent != "" {
		htmlTemplate, err := template.New("html").Parse(tmpl.HTMLContent)
		if err != nil {
			return "", "", fmt.Errorf("failed to parse HTML template: %w", err)
		}

		var htmlBuffer bytes.Buffer
		if err := htmlTemplate.Execute(&htmlBuffer, params); err != nil {
			return "", "", fmt.Errorf("failed to execute HTML template: %w", err)
		}
		htmlContent = htmlBuffer.String()
	}

	// Process text content
	var textContent string
	if tmpl.TextContent != "" {
		textTemplate, err := template.New("text").Parse(tmpl.TextContent)
		if err != nil {
			return "", "", fmt.Errorf("failed to parse text template: %w", err)
		}

		var textBuffer bytes.Buffer
		if err := textTemplate.Execute(&textBuffer, params); err != nil {
			return "", "", fmt.Errorf("failed to execute text template: %w", err)
		}
		textContent = textBuffer.String()
	} else if htmlContent != "" {
		// Generate simple text version from HTML if no text template exists
		textContent = stripHTML(htmlContent)
	}

	return htmlContent, textContent, nil
}

// sendEmailViaSMTP sends email via SMTP
func (s *EmailService) sendEmailViaSMTP(service models.EmailService, emailLog models.EmailLog, htmlContent, textContent string, attachments []models.EmailAttachment) error {
	// Parse configuration
	config := parseSMTPConfig(service.Configuration)

	log.Printf("SMTP Config - Host: %s, Port: %d, Username: %s, UseTLS: %v, UseSSL: %v",
		config.Host, config.Port, config.Username, config.UseTLS, config.UseSSL)

	// Validate configuration
	if config.Host == "" || config.Port == 0 {
		return fmt.Errorf("invalid SMTP configuration: missing host or port")
	}

	// Create SMTP auth
	auth := smtp.PlainAuth("", config.Username, config.Password, config.Host)

	// Setup TLS config
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         config.Host,
	}

	// Determine the connection method based on configuration
	var client *smtp.Client
	var err error

	if config.UseSSL {
		// SSL/TLS connection (port 465)
		log.Printf("Connecting via SSL to %s:%d", config.Host, config.Port)
		conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", config.Host, config.Port), tlsConfig)
		if err != nil {
			return fmt.Errorf("failed to connect to SMTP server (SSL): %w", err)
		}
		defer conn.Close()

		client, err = smtp.NewClient(conn, config.Host)
		if err != nil {
			return fmt.Errorf("failed to create SMTP client: %w", err)
		}
	} else {
		// Plain or STARTTLS connection (port 587/25)
		log.Printf("Connecting to %s:%d", config.Host, config.Port)
		client, err = smtp.Dial(fmt.Sprintf("%s:%d", config.Host, config.Port))
		if err != nil {
			return fmt.Errorf("failed to connect to SMTP server: %w", err)
		}

		// Use STARTTLS if enabled
		if config.UseTLS {
			log.Printf("Starting TLS negotiation")
			if err = client.StartTLS(tlsConfig); err != nil {
				return fmt.Errorf("failed to start TLS: %w", err)
			}
		}
	}
	defer client.Quit()

	// Authenticate
	if config.Username != "" && config.Password != "" {
		log.Printf("Authenticating as %s", config.Username)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP authentication failed: %w", err)
		}
		log.Printf("Authentication successful")
	}

	// Set sender - use from_email from config if available
	fromEmail := emailLog.FromEmail
	if config.FromEmail != "" {
		fromEmail = config.FromEmail
	}
	log.Printf("Setting sender: %s", fromEmail)
	if err := client.Mail(fromEmail); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// Set recipient
	log.Printf("Setting recipient: %s", emailLog.ToEmail)
	if err := client.Rcpt(emailLog.ToEmail); err != nil {
		return fmt.Errorf("failed to set recipient: %w", err)
	}

	// Compose message
	message := composeMessage(emailLog, htmlContent, textContent, attachments)
	log.Printf("Composed message: %d bytes", len(message))

	// Send message
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to initialize data: %w", err)
	}
	defer w.Close()

	_, err = w.Write([]byte(message))
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	log.Printf("✓ Email sent successfully to %s", emailLog.ToEmail)
	return nil
}

// Helper functions
func getFromEmail(reqFromEmail string, service models.EmailService) string {
	if reqFromEmail != "" {
		return reqFromEmail
	}
	// Parse config to get from_email
	config := parseSMTPConfig(service.Configuration)
	if config.FromEmail != "" {
		return config.FromEmail
	}
	return "noreply@example.com"
}

func getFromName(reqFromName string, service models.EmailService) string {
	if reqFromName != "" {
		return reqFromName
	}
	// Parse config to get from_name if available
	var configMap map[string]interface{}
	if err := json.Unmarshal([]byte(service.Configuration), &configMap); err == nil {
		if fromName, ok := configMap["from_name"].(string); ok && fromName != "" {
			return fromName
		}
	}
	return "LeapMailr"
}

func stripHTML(html string) string {
	// Simple HTML stripper - in production use a proper library
	text := strings.ReplaceAll(html, "<br>", "\n")
	text = strings.ReplaceAll(text, "<br/>", "\n")
	text = strings.ReplaceAll(text, "<br />", "\n")
	// TODO: Implement proper HTML stripping
	return text
}

func parseSMTPConfig(config string) struct {
	Host      string
	Port      int
	Username  string
	Password  string
	FromEmail string
	UseTLS    bool
	UseSSL    bool
} {
	// Parse JSON configuration
	var configMap map[string]interface{}
	if err := json.Unmarshal([]byte(config), &configMap); err != nil {
		// Return default if parsing fails
		return struct {
			Host      string
			Port      int
			Username  string
			Password  string
			FromEmail string
			UseTLS    bool
			UseSSL    bool
		}{
			Host:      "smtp.example.com",
			Port:      587,
			Username:  "user@example.com",
			Password:  "password",
			FromEmail: "noreply@example.com",
			UseTLS:    true,
			UseSSL:    false,
		}
	}

	// Extract values with type assertions and defaults
	host := ""
	if h, ok := configMap["host"].(string); ok {
		host = h
	}

	port := 587
	if p, ok := configMap["port"].(float64); ok {
		port = int(p)
	}

	username := ""
	if u, ok := configMap["username"].(string); ok {
		username = u
	}

	password := ""
	if p, ok := configMap["password"].(string); ok {
		password = p
	}

	fromEmail := ""
	if f, ok := configMap["from_email"].(string); ok {
		fromEmail = f
	}

	useTLS := true
	if t, ok := configMap["use_tls"].(bool); ok {
		useTLS = t
	}

	useSSL := false
	if s, ok := configMap["use_ssl"].(bool); ok {
		useSSL = s
	}

	return struct {
		Host      string
		Port      int
		Username  string
		Password  string
		FromEmail string
		UseTLS    bool
		UseSSL    bool
	}{
		Host:      host,
		Port:      port,
		Username:  username,
		Password:  password,
		FromEmail: fromEmail,
		UseTLS:    useTLS,
		UseSSL:    useSSL,
	}
}

func composeMessage(log models.EmailLog, htmlContent, textContent string, attachments []models.EmailAttachment) string {
	var message bytes.Buffer

	// Headers
	message.WriteString(fmt.Sprintf("From: %s <%s>\r\n", log.FromName, log.FromEmail))
	message.WriteString(fmt.Sprintf("To: %s <%s>\r\n", log.ToName, log.ToEmail))
	message.WriteString(fmt.Sprintf("Subject: %s\r\n", log.Subject))
	message.WriteString("MIME-Version: 1.0\r\n")

	if len(attachments) > 0 || (htmlContent != "" && textContent != "") {
		// Multipart message
		boundary := "boundary-" + uuid.New().String()
		message.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\r\n\r\n", boundary))

		// Text part
		if textContent != "" {
			message.WriteString(fmt.Sprintf("--%s\r\n", boundary))
			message.WriteString("Content-Type: text/plain; charset=utf-8\r\n\r\n")
			message.WriteString(textContent)
			message.WriteString("\r\n\r\n")
		}

		// HTML part
		if htmlContent != "" {
			message.WriteString(fmt.Sprintf("--%s\r\n", boundary))
			message.WriteString("Content-Type: text/html; charset=utf-8\r\n\r\n")
			message.WriteString(htmlContent)
			message.WriteString("\r\n\r\n")
		}

		// Attachments
		for _, attachment := range attachments {
			message.WriteString(fmt.Sprintf("--%s\r\n", boundary))
			message.WriteString(fmt.Sprintf("Content-Type: %s\r\n", attachment.ContentType))
			message.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n", attachment.Filename))
			message.WriteString("Content-Transfer-Encoding: base64\r\n\r\n")
			// TODO: Implement base64 encoding of attachment content
			message.WriteString("\r\n")
		}

		message.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
	} else {
		// Simple message
		if htmlContent != "" {
			message.WriteString("Content-Type: text/html; charset=utf-8\r\n\r\n")
			message.WriteString(htmlContent)
		} else {
			message.WriteString("Content-Type: text/plain; charset=utf-8\r\n\r\n")
			message.WriteString(textContent)
		}
	}

	return message.String()
}

// GetEmailHistory retrieves email history for a user
func (s *EmailService) GetEmailHistory(userID uuid.UUID, emailLogs *[]models.EmailLog) error {
	return s.db.Where("user_id = ? OR organization_id IN (SELECT organization_id FROM user_organizations WHERE user_id = ?)",
		userID, userID).
		Order("created_at DESC").
		Limit(100).
		Find(emailLogs).Error
}

// GetEmailStatus retrieves the status of a specific email
func (s *EmailService) GetEmailStatus(emailID, userID uuid.UUID) (*models.EmailLog, error) {
	var emailLog models.EmailLog
	err := s.db.Where("id = ? AND (user_id = ? OR organization_id IN (SELECT organization_id FROM user_organizations WHERE user_id = ?))",
		emailID, userID, userID).First(&emailLog).Error
	if err != nil {
		return nil, err
	}
	return &emailLog, nil
}
