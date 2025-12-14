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
	"github.com/dhawalhost/leapmailr/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Email service constants
const (
	defaultLocalhost   = "leapmailr.local"
	defaultNoReplyAddr = "noreply@example.com"
)

// smtpConfig holds SMTP configuration
type smtpConfig struct {
	Host      string
	Port      int
	Username  string
	Password  string
	FromEmail string
	UseTLS    bool
	UseSSL    bool
}

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
	if err := s.checkSuppression(req.ToEmail, userID); err != nil {
		return nil, err
	}

	// Get template and service
	emailTemplate, emailService, err := s.getTemplateAndService(req.TemplateID, req.ServiceID, userID)
	if err != nil {
		return nil, err
	}

	// Process template content
	htmlContent, textContent, subject, err := s.prepareEmailContent(emailTemplate, req)
	if err != nil {
		return nil, err
	}

	// Create email log entry
	emailLog, err := s.createEmailLog(emailTemplate, emailService, req, userID, subject, htmlContent, textContent)
	if err != nil {
		return nil, err
	}

	// Handle scheduled emails
	if s.isScheduled(req.ScheduleAt) {
		return s.scheduleEmail(emailLog)
	}

	// Send email immediately
	if err := s.sendImmediately(emailService, emailTemplate, emailLog, htmlContent, textContent, req.Attachments); err != nil {
		return emailLog, err
	}

	// Send auto-reply if enabled
	s.handleAutoReply(emailTemplate, emailService, req, userID)

	return emailLog, nil
}

// checkSuppression verifies the recipient is not suppressed
func (s *EmailService) checkSuppression(toEmail string, userID uuid.UUID) error {
	suppressionService := NewSuppressionService()
	isSuppressed, suppression, err := suppressionService.IsEmailSuppressed(toEmail, userID)
	if err != nil {
		return fmt.Errorf("failed to check suppression list: %w", err)
	}
	if isSuppressed {
		return fmt.Errorf("email address is suppressed (reason: %s)", suppression.Reason)
	}
	return nil
}

// getTemplateAndService retrieves the template and email service
func (s *EmailService) getTemplateAndService(templateID uuid.UUID, serviceID *uuid.UUID, userID uuid.UUID) (models.Template, models.EmailService, error) {
	var emailTemplate models.Template
	if err := s.db.Where("id = ? AND (user_id = ? OR organization_id IN (SELECT organization_id FROM user_organizations WHERE user_id = ?))",
		templateID, userID, userID).First(&emailTemplate).Error; err != nil {
		return emailTemplate, models.EmailService{}, fmt.Errorf("template not found: %w", err)
	}

	var emailService models.EmailService
	if serviceID != nil {
		if err := s.db.Where("id = ? AND (user_id = ? OR organization_id IN (SELECT organization_id FROM user_organizations WHERE user_id = ?))",
			*serviceID, userID, userID).First(&emailService).Error; err != nil {
			return emailTemplate, emailService, fmt.Errorf("email service not found: %w", err)
		}
	} else {
		// Use default service
		if err := s.db.Where("(user_id = ? OR organization_id IN (SELECT organization_id FROM user_organizations WHERE user_id = ?)) AND is_default = ?",
			userID, userID, true).First(&emailService).Error; err != nil {
			return emailTemplate, emailService, fmt.Errorf("no default email service found: %w", err)
		}
	}

	return emailTemplate, emailService, nil
}

// prepareEmailContent processes the template and returns HTML, text, and subject
func (s *EmailService) prepareEmailContent(emailTemplate models.Template, req models.EmailRequest) (string, string, string, error) {
	subject := req.Subject
	if subject == "" {
		subject = emailTemplate.Subject
	}

	htmlContent, textContent, err := s.processTemplate(emailTemplate, req.TemplateParams, subject)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to process template: %w", err)
	}

	return htmlContent, textContent, subject, nil
}

// createEmailLog creates an email log entry and applies tracking if enabled
func (s *EmailService) createEmailLog(emailTemplate models.Template, emailService models.EmailService, req models.EmailRequest, userID uuid.UUID, subject, htmlContent, textContent string) (*models.EmailLog, error) {
	fromEmail := getFromEmailWithTemplate(req.FromEmail, emailTemplate, emailService)
	domain := extractDomain(fromEmail)

	emailLog := &models.EmailLog{
		UserID:     &userID,
		TemplateID: &emailTemplate.ID,
		ServiceID:  &emailService.ID,
		FromEmail:  fromEmail,
		FromName:   getFromNameWithTemplate(req.FromName, emailTemplate, emailService),
		ToEmail:    req.ToEmail,
		ToName:     req.ToName,
		Subject:    subject,
		Status:     "queued",
		MessageID:  generateMessageID(domain),
		Metadata:   "{}",
	}

	if err := s.db.Create(emailLog).Error; err != nil {
		return nil, fmt.Errorf("failed to create email log: %w", err)
	}

	// Apply tracking if enabled
	if req.EnableTracking && htmlContent != "" {
		s.applyTracking(emailLog.ID, &htmlContent)
	}

	return emailLog, nil
}

// applyTracking injects tracking pixel and link tracking into HTML content
func (s *EmailService) applyTracking(emailLogID uuid.UUID, htmlContent *string) {
	trackingService := NewEmailTrackingService()
	tracking, err := trackingService.CreateTracking(emailLogID)
	if err == nil && tracking != nil {
		baseURL := utils.GetBaseURL()
		*htmlContent = trackingService.InjectTrackingPixel(*htmlContent, tracking.TrackingPixelID, baseURL)
		*htmlContent = trackingService.InjectLinkTracking(*htmlContent, tracking.TrackingPixelID, baseURL)
	}
}

// extractDomain extracts domain from email address
func extractDomain(email string) string {
	if parts := strings.Split(email, "@"); len(parts) == 2 {
		return parts[1]
	}
	return defaultLocalhost
}

// isScheduled checks if the email should be scheduled
func (s *EmailService) isScheduled(scheduleAt *time.Time) bool {
	return scheduleAt != nil && scheduleAt.After(time.Now())
}

// scheduleEmail marks email as scheduled
func (s *EmailService) scheduleEmail(emailLog *models.EmailLog) (*models.EmailLog, error) {
	emailLog.Status = "scheduled"
	s.db.Save(emailLog)
	// TODO: Add to job queue for scheduled sending
	return emailLog, nil
}

// sendImmediately sends the email via SMTP immediately
func (s *EmailService) sendImmediately(emailService models.EmailService, emailTemplate models.Template, emailLog *models.EmailLog, htmlContent, textContent string, attachments []models.EmailAttachment) error {
	log.Printf("Attempting to send email to %s via service %s", emailLog.ToEmail, emailService.Name)

	err := s.sendEmailViaSMTP(emailService, emailTemplate, *emailLog, htmlContent, textContent, attachments)
	if err != nil {
		log.Printf("Failed to send email: %v", err)
		emailLog.Status = "failed"
		emailLog.ErrorMessage = err.Error()
		s.db.Save(emailLog)
		return err
	}

	log.Printf("Email sent successfully to %s", emailLog.ToEmail)
	emailLog.Status = "sent"
	now := time.Now()
	emailLog.SentAt = &now
	s.db.Save(emailLog)

	return nil
}

// handleAutoReply sends auto-reply if enabled
func (s *EmailService) handleAutoReply(emailTemplate models.Template, emailService models.EmailService, req models.EmailRequest, userID uuid.UUID) {
	autoReplyTemplateID := emailTemplate.AutoReplyTemplateID
	if req.AutoReplyEnabled && req.AutoReplyTemplateID != nil {
		// Request-level override
		autoReplyTemplateID = req.AutoReplyTemplateID
	}

	if emailTemplate.AutoReplyEnabled && autoReplyTemplateID != nil {
		log.Printf("Auto-reply enabled for template %s, sending auto-reply using template %s", emailTemplate.Name, *autoReplyTemplateID)
		go s.sendAutoReply(emailService, *autoReplyTemplateID, req.ToEmail, req.ToName, userID)
	}
}

// generateMessageID generates a unique Message-ID for email tracking
// Format: <uuid@domain> which is RFC-compliant
func generateMessageID(domain string) string {
	if domain == "" {
		domain = defaultLocalhost
	}
	return fmt.Sprintf("%s@%s", uuid.New().String(), domain)
}

// SendBulkEmail sends bulk emails
func (s *EmailService) SendBulkEmail(req models.BulkEmailRequest, userID uuid.UUID) ([]models.EmailLog, error) {
	// Get template and service
	emailTemplate, emailService, err := s.getTemplateAndService(req.TemplateID, req.ServiceID, userID)
	if err != nil {
		return nil, err
	}

	// Process each recipient
	return s.processRecipients(req, emailTemplate, emailService, userID)
}

// processRecipients processes all recipients and sends bulk emails
func (s *EmailService) processRecipients(req models.BulkEmailRequest, emailTemplate models.Template, emailService models.EmailService, userID uuid.UUID) ([]models.EmailLog, error) {
	var emailLogs []models.EmailLog
	suppressionService := NewSuppressionService()

	for _, recipient := range req.Recipients {
		// Check suppression list
		if s.isRecipientSuppressed(recipient.Email, userID, suppressionService) {
			continue
		}

		// Process recipient's email
		emailLog, htmlContent, textContent, err := s.processRecipient(req, recipient, emailTemplate, emailService, userID)
		if err != nil {
			continue // Skip this recipient on error
		}

		emailLogs = append(emailLogs, *emailLog)

		// Send email (scheduled or immediate)
		if s.isScheduled(req.ScheduleAt) {
			s.scheduleBulkEmail(emailLog)
		} else {
			s.sendBulkEmailAsync(emailLog, emailService, emailTemplate, htmlContent, textContent, req.EnableTracking)
		}
	}

	return emailLogs, nil
}

// isRecipientSuppressed checks if recipient is in suppression list
func (s *EmailService) isRecipientSuppressed(email string, userID uuid.UUID, suppressionService *SuppressionService) bool {
	isSuppressed, _, err := suppressionService.IsEmailSuppressed(email, userID)
	return err != nil || isSuppressed
}

// processRecipient processes a single recipient and creates email log
func (s *EmailService) processRecipient(req models.BulkEmailRequest, recipient models.EmailRecipient, emailTemplate models.Template, emailService models.EmailService, userID uuid.UUID) (*models.EmailLog, string, string, error) {
	// Merge template params
	templateParams := mergeTemplateParams(req.DefaultParams, recipient.TemplateParams)

	// Determine subject
	subject := req.Subject
	if subject == "" {
		subject = emailTemplate.Subject
	}

	// Process template
	htmlContent, textContent, err := s.processTemplate(emailTemplate, templateParams, subject)
	if err != nil {
		return nil, "", "", err
	}

	// Create email log
	fromEmail := getFromEmailWithTemplate(req.FromEmail, emailTemplate, emailService)
	domain := extractDomain(fromEmail)

	emailLog := &models.EmailLog{
		UserID:     &userID,
		TemplateID: &emailTemplate.ID,
		ServiceID:  &emailService.ID,
		FromEmail:  fromEmail,
		FromName:   getFromNameWithTemplate(req.FromName, emailTemplate, emailService),
		ToEmail:    recipient.Email,
		ToName:     recipient.Name,
		Subject:    subject,
		Status:     "queued",
		MessageID:  generateMessageID(domain),
		Metadata:   "{}",
	}

	if err := s.db.Create(emailLog).Error; err != nil {
		return nil, "", "", err
	}

	return emailLog, htmlContent, textContent, nil
}

// mergeTemplateParams merges default and recipient-specific template params
func mergeTemplateParams(defaultParams, recipientParams map[string]interface{}) map[string]interface{} {
	templateParams := make(map[string]interface{})
	for k, v := range defaultParams {
		templateParams[k] = v
	}
	for k, v := range recipientParams {
		templateParams[k] = v
	}
	return templateParams
}

// scheduleBulkEmail marks a bulk email as scheduled
func (s *EmailService) scheduleBulkEmail(emailLog *models.EmailLog) {
	emailLog.Status = "scheduled"
	s.db.Save(emailLog)
	// TODO: Add to job queue
}

// sendBulkEmailAsync sends a bulk email asynchronously
func (s *EmailService) sendBulkEmailAsync(emailLog *models.EmailLog, emailService models.EmailService, emailTemplate models.Template, htmlContent, textContent string, enableTracking bool) {
	go func(log models.EmailLog, html, text string, tracking bool) {
		// Apply tracking if enabled
		if tracking && html != "" {
			s.applyTracking(log.ID, &html)
		}

		// Send email
		err := s.sendEmailViaSMTP(emailService, emailTemplate, log, html, text, nil)
		if err != nil {
			log.Status = "failed"
			log.ErrorMessage = err.Error()
		} else {
			log.Status = "sent"
			now := time.Now()
			log.SentAt = &now
		}
		s.db.Save(&log)
	}(*emailLog, htmlContent, textContent, enableTracking)
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
	}
	// Note: We don't auto-generate text from HTML anymore
	// Only send HTML version for better formatting

	return htmlContent, textContent, nil
}

// sendEmailViaSMTP sends email via SMTP
func (s *EmailService) sendEmailViaSMTP(service models.EmailService, template models.Template, emailLog models.EmailLog, htmlContent, textContent string, attachments []models.EmailAttachment) error {
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

	// Connect to SMTP server
	client, err := s.connectSMTPClient(config, tlsConfig)
	if err != nil {
		return err
	}
	defer func() { _ = client.Quit() }()

	// Authenticate
	if err := s.authenticateSMTP(client, auth, config); err != nil {
		return err
	}

	// Send the email
	if err := s.sendSMTPMessage(client, emailLog, template, service, htmlContent, textContent, attachments); err != nil {
		return err
	}

	log.Printf("✓ Email sent successfully to %s", emailLog.ToEmail)
	return nil
}

func (s *EmailService) connectSMTPClient(config smtpConfig, tlsConfig *tls.Config) (*smtp.Client, error) {
	var client *smtp.Client
	var err error

	if config.UseSSL {
		// SSL/TLS connection (port 465)
		log.Printf("Connecting via SSL to %s:%d", config.Host, config.Port)
		conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", config.Host, config.Port), tlsConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to SMTP server (SSL): %w", err)
		}

		client, err = smtp.NewClient(conn, config.Host)
		if err != nil {
			conn.Close()
			return nil, fmt.Errorf("failed to create SMTP client: %w", err)
		}
	} else {
		// Plain or STARTTLS connection (port 587/25)
		log.Printf("Connecting to %s:%d", config.Host, config.Port)
		client, err = smtp.Dial(fmt.Sprintf("%s:%d", config.Host, config.Port))
		if err != nil {
			return nil, fmt.Errorf("failed to connect to SMTP server: %w", err)
		}

		// Use STARTTLS if enabled
		if config.UseTLS {
			log.Printf("Starting TLS negotiation")
			if err = client.StartTLS(tlsConfig); err != nil {
				client.Close()
				return nil, fmt.Errorf("failed to start TLS: %w", err)
			}
		}
	}

	return client, nil
}

func (s *EmailService) authenticateSMTP(client *smtp.Client, auth smtp.Auth, config smtpConfig) error {
	if config.Username != "" && config.Password != "" {
		log.Printf("Authenticating as %s", config.Username)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP authentication failed: %w", err)
		}
		log.Printf("Authentication successful")
	}
	return nil
}

func (s *EmailService) sendSMTPMessage(client *smtp.Client, emailLog models.EmailLog, template models.Template, service models.EmailService, htmlContent, textContent string, attachments []models.EmailAttachment) error {
	// Set sender
	fromEmail := emailLog.FromEmail
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
	message := composeMessage(template, service, emailLog, htmlContent, textContent, attachments)
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

	return nil
}

// sendAutoReply sends an automated reply email
func (s *EmailService) sendAutoReply(service models.EmailService, autoReplyTemplateID uuid.UUID, toEmail, toName string, userID uuid.UUID) {
	// Get the auto-reply template
	var autoReplyTemplate models.Template
	if err := s.db.Where("id = ?", autoReplyTemplateID).First(&autoReplyTemplate).Error; err != nil {
		log.Printf("Failed to load auto-reply template %s: %v", autoReplyTemplateID, err)
		return
	}

	// Prepare auto-reply parameters with common variables
	autoReplyParams := map[string]interface{}{
		"name":            toName,
		"email":           toEmail,
		"reference_id":    uuid.New().String()[:8], // Short reference ID
		"received_date":   time.Now().Format("January 2, 2006 at 3:04 PM"),
		"response_time":   "24-48",           // Default response time
		"company_name":    "Your Company",    // Can be customized per service
		"support_phone":   "+1-234-567-8900", // Can be customized
		"website_url":     "https://yourcompany.com",
		"help_center_url": "https://yourcompany.com/help",
		"year":            time.Now().Year(),
	}

	// Process the auto-reply template
	htmlContent, textContent, err := s.processTemplate(autoReplyTemplate, autoReplyParams, autoReplyTemplate.Subject)
	if err != nil {
		log.Printf("Failed to process auto-reply template: %v", err)
		return
	}

	// Create email log for auto-reply
	fromEmail := getFromEmailWithTemplate("", autoReplyTemplate, service)
	domain := defaultLocalhost
	if parts := strings.Split(fromEmail, "@"); len(parts) == 2 {
		domain = parts[1]
	}

	emailLog := models.EmailLog{
		UserID:     &userID,
		TemplateID: &autoReplyTemplate.ID,
		ServiceID:  &service.ID,
		FromEmail:  fromEmail,
		FromName:   getFromNameWithTemplate("", autoReplyTemplate, service),
		ToEmail:    toEmail,
		ToName:     toName,
		Subject:    autoReplyTemplate.Subject,
		Status:     "queued",
		MessageID:  generateMessageID(domain),
		Metadata:   `{"auto_reply": true}`,
	}

	if err := s.db.Create(&emailLog).Error; err != nil {
		log.Printf("Failed to create email log for auto-reply: %v", err)
		return
	}

	// Send the auto-reply email
	log.Printf("Sending auto-reply to %s", toEmail)
	err = s.sendEmailViaSMTP(service, autoReplyTemplate, emailLog, htmlContent, textContent, nil)
	if err != nil {
		log.Printf("Failed to send auto-reply: %v", err)
		emailLog.Status = "failed"
		emailLog.ErrorMessage = err.Error()
		s.db.Save(&emailLog)
		return
	}

	log.Printf("Auto-reply sent successfully to %s", toEmail)
	emailLog.Status = "sent"
	now := time.Now()
	emailLog.SentAt = &now
	s.db.Save(&emailLog)
}

// Helper functions
// Priority: Request override > Template override > Service default > Fallback
func getFromEmailWithTemplate(reqFromEmail string, template models.Template, service models.EmailService) string {
	// 1st priority: Request override (for one-off changes)
	if reqFromEmail != "" {
		return reqFromEmail
	}
	// 2nd priority: Template override (configured in template)
	if template.FromEmail != "" {
		return template.FromEmail
	}
	// 3rd priority: Service's configured sender email
	if service.FromEmail != "" {
		return service.FromEmail
	}
	// Fallback
	return defaultNoReplyAddr
}

func getFromNameWithTemplate(reqFromName string, template models.Template, service models.EmailService) string {
	// 1st priority: Request override (for one-off changes)
	if reqFromName != "" {
		return reqFromName
	}
	// 2nd priority: Template override (configured in template)
	if template.FromName != "" {
		return template.FromName
	}
	// 3rd priority: Service's configured sender name
	if service.FromName != "" {
		return service.FromName
	}
	// Fallback
	return "LeapMailr"
}

func getReplyToEmail(template models.Template, service models.EmailService) string {
	// 1st priority: Template override
	if template.ReplyToEmail != "" {
		return template.ReplyToEmail
	}
	// 2nd priority: Service's configured reply-to
	if service.ReplyToEmail != "" {
		return service.ReplyToEmail
	}
	// Use from_email if no reply-to is set
	if service.FromEmail != "" {
		return service.FromEmail
	}
	return ""
}

func parseSMTPConfig(config string) smtpConfig {
	// Decrypt configuration (GAP-SEC-005)
	encryption, err := utils.NewEncryptionService()
	var decryptedConfig string
	if err == nil {
		decryptedConfig, err = encryption.Decrypt(config)
		if err != nil {
			// If decryption fails, try to parse as unencrypted (for backward compatibility)
			decryptedConfig = config
		}
	} else {
		decryptedConfig = config
	}

	// Parse JSON configuration
	var configMap map[string]interface{}
	if err := json.Unmarshal([]byte(decryptedConfig), &configMap); err != nil {
		// Return default if parsing fails
		return smtpConfig{
			Host:      "smtp.example.com",
			Port:      587,
			Username:  "user@example.com",
			Password:  "password",
			FromEmail: defaultNoReplyAddr,
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

	return smtpConfig{
		Host:      host,
		Port:      port,
		Username:  username,
		Password:  password,
		FromEmail: fromEmail,
		UseTLS:    useTLS,
		UseSSL:    useSSL,
	}
}

func composeMessage(template models.Template, service models.EmailService, log models.EmailLog, htmlContent, textContent string, attachments []models.EmailAttachment) string {
	var message bytes.Buffer

	// Headers
	message.WriteString(fmt.Sprintf("From: %s <%s>\r\n", log.FromName, log.FromEmail))
	message.WriteString(fmt.Sprintf("To: %s <%s>\r\n", log.ToName, log.ToEmail))

	// Add unique Message-ID header for tracking
	if log.MessageID != "" {
		message.WriteString(fmt.Sprintf("Message-ID: <%s>\r\n", log.MessageID))
	}

	// Add Reply-To header if configured (template override > service default)
	replyTo := getReplyToEmail(template, service)
	if replyTo != "" {
		message.WriteString(fmt.Sprintf("Reply-To: %s\r\n", replyTo))
	}

	message.WriteString(fmt.Sprintf("Subject: %s\r\n", log.Subject))
	message.WriteString("MIME-Version: 1.0\r\n")

	if len(attachments) > 0 {
		// Multipart message with attachments
		boundary := "boundary-" + uuid.New().String()
		message.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\r\n\r\n", boundary))

		// HTML part (prefer HTML over text)
		if htmlContent != "" {
			message.WriteString(fmt.Sprintf("--%s\r\n", boundary))
			message.WriteString("Content-Type: text/html; charset=utf-8\r\n\r\n")
			message.WriteString(htmlContent)
			message.WriteString("\r\n\r\n")
		} else if textContent != "" {
			message.WriteString(fmt.Sprintf("--%s\r\n", boundary))
			message.WriteString("Content-Type: text/plain; charset=utf-8\r\n\r\n")
			message.WriteString(textContent)
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
		// Simple message - send only HTML or only text, not both
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
