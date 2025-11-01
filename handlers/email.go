package handlers

import (
	"net/http"

	"github.com/dhawalhost/leapmailr/models"
	"github.com/dhawalhost/leapmailr/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// SendEmailHandler handles single email sending via new API
func SendEmailHandler(c *gin.Context) {
	user, err := GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}

	var req models.EmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	emailService := service.NewEmailService()
	emailLog, err := emailService.SendEmail(req, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to send email",
			"details": err.Error(),
		})
		return
	}

	// Check for and send auto-reply if configured
	if emailLog.Status == "sent" && req.ToEmail != "" {
		// Try to get active auto-reply for this service
		serviceIDForAutoReply := uuid.Nil
		if req.ServiceID != nil {
			serviceIDForAutoReply = *req.ServiceID
		}

		autoReplyConfig, err := service.GetActiveAutoReplyForService(user.ID, serviceIDForAutoReply, false) // false = API send, not form
		if err == nil && autoReplyConfig != nil {
			// Send auto-reply asynchronously
			go func() {
				// Convert template params to string map
				variables := make(map[string]string)
				for k, v := range req.TemplateParams {
					if str, ok := v.(string); ok {
						variables[k] = str
					}
				}
				service.SendAutoReply(autoReplyConfig, req.ToEmail, variables)
			}()
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Email sent successfully",
		"data": gin.H{
			"email_id": emailLog.ID,
			"status":   emailLog.Status,
		},
	})
}

// SendBulkEmailHandler handles bulk email sending
func SendBulkEmailHandler(c *gin.Context) {
	user, err := GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}

	var req models.BulkEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	emailService := service.NewEmailService()
	emailLogs, err := emailService.SendBulkEmail(req, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to send bulk emails",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Bulk emails queued successfully",
		"data": gin.H{
			"total_emails": len(emailLogs),
			"email_ids":    extractEmailIDs(emailLogs),
		},
	})
}

// SendFormHandler handles form-based email sending (EmailJS compatibility)
func SendFormHandler(c *gin.Context) {
	// This endpoint maintains compatibility with EmailJS-style requests
	var formData map[string]interface{}
	if err := c.ShouldBindJSON(&formData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid form data",
		})
		return
	}

	// Extract required fields
	serviceID, _ := formData["service_id"].(string)
	templateID, _ := formData["template_id"].(string)
	userID, _ := formData["user_id"].(string) // This is actually the API key
	templateParams, _ := formData["template_params"].(map[string]interface{})
	captchaToken, _ := formData["captcha_token"].(string)
	captchaConfigID, _ := formData["captcha_config_id"].(string)

	if templateID == "" || userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing required parameters",
		})
		return
	}

	// Validate CAPTCHA if provided
	if captchaToken != "" && captchaConfigID != "" {
		captchaService := service.NewCaptchaService()
		configUUID := parseUUID(captchaConfigID)

		isValid, err := captchaService.VerifyCaptcha(captchaToken, configUUID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "CAPTCHA validation failed: " + err.Error(),
			})
			return
		}

		if !isValid {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid CAPTCHA token",
			})
			return
		}
	}

	// Validate API key
	authService := service.NewAuthService()
	user, err := authService.ValidateAPIKey(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid API key",
		})
		return
	}

	// Build email request
	emailReq := models.EmailRequest{
		TemplateID:     parseUUID(templateID),
		TemplateParams: templateParams,
	}

	if serviceID != "" && serviceID != "default_service" {
		serviceUUID := parseUUID(serviceID)
		emailReq.ServiceID = &serviceUUID
	}

	// Extract recipient from template params (EmailJS style)
	if email, ok := templateParams["to_email"].(string); ok {
		emailReq.ToEmail = email
	}
	if name, ok := templateParams["to_name"].(string); ok {
		emailReq.ToName = name
	}

	emailService := service.NewEmailService()
	emailLog, err := emailService.SendEmail(emailReq, user.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Check for and send auto-reply if configured
	if emailLog.Status == "sent" && emailReq.ToEmail != "" {
		// Try to get active auto-reply for this service
		serviceIDForAutoReply := uuid.Nil
		if emailReq.ServiceID != nil {
			serviceIDForAutoReply = *emailReq.ServiceID
		}

		autoReplyConfig, err := service.GetActiveAutoReplyForService(user.ID, serviceIDForAutoReply, true)
		if err == nil && autoReplyConfig != nil {
			// Send auto-reply asynchronously (don't block the response)
			go func() {
				// Use the template params as variables for auto-reply
				variables := make(map[string]string)
				for k, v := range templateParams {
					if str, ok := v.(string); ok {
						variables[k] = str
					}
				}
				service.SendAutoReply(autoReplyConfig, emailReq.ToEmail, variables)
			}()
		}

		// Collect contact if enabled (check for collect_contact flag, default true for backward compatibility)
		collectContact := true
		if collectFlag, ok := formData["collect_contact"].(bool); ok {
			collectContact = collectFlag
		}

		if collectContact && emailReq.ToEmail != "" {
			// Extract contact information from template params
			go func() {
				contactReq := models.CreateContactRequest{
					Email:  emailReq.ToEmail,
					Source: "form",
				}

				if name, ok := templateParams["to_name"].(string); ok {
					contactReq.Name = name
				} else if name, ok := templateParams["from_name"].(string); ok {
					contactReq.Name = name
				} else if name, ok := templateParams["name"].(string); ok {
					contactReq.Name = name
				}

				if phone, ok := templateParams["phone"].(string); ok {
					contactReq.Phone = phone
				}

				if company, ok := templateParams["company"].(string); ok {
					contactReq.Company = company
				}

				// Store all template params as metadata
				metadata := make(map[string]string)
				for k, v := range templateParams {
					if k != "to_email" && k != "to_name" && k != "from_name" && k != "name" && k != "phone" && k != "company" {
						if str, ok := v.(string); ok {
							metadata[k] = str
						}
					}
				}
				if len(metadata) > 0 {
					contactReq.Metadata = metadata
				}

				// Tag with template ID
				contactReq.Tags = []string{templateID}

				service.CreateContact(contactReq, user.ID)
			}()
		}
	}

	// EmailJS-compatible response
	if emailLog.Status == "sent" {
		c.JSON(http.StatusOK, "OK")
	} else {
		c.JSON(http.StatusBadRequest, "Failed to send email")
	}
}

// GetEmailHistoryHandler returns email sending history
func GetEmailHistoryHandler(c *gin.Context) {
	user, err := GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}

	// TODO: Implement pagination and filtering
	var emailLogs []models.EmailLog
	emailService := service.NewEmailService()
	if err := emailService.GetEmailHistory(user.ID, &emailLogs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve email history",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   emailLogs,
	})
}

// GetEmailStatusHandler returns the status of a specific email
func GetEmailStatusHandler(c *gin.Context) {
	user, err := GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}

	emailID := c.Param("id")
	if emailID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Email ID is required",
		})
		return
	}

	emailService := service.NewEmailService()
	emailLog, err := emailService.GetEmailStatus(parseUUID(emailID), user.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Email not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   emailLog,
	})
}

// Helper functions
func extractEmailIDs(logs []models.EmailLog) []string {
	ids := make([]string, len(logs))
	for i, log := range logs {
		ids[i] = log.ID.String()
	}
	return ids
}

func parseUUID(str string) uuid.UUID {
	id, _ := uuid.Parse(str)
	return id
}
