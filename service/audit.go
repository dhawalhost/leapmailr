package service

import (
	"encoding/json"
	"time"

	"github.com/dhawalhost/leapmailr/database"
	"github.com/dhawalhost/leapmailr/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuditService handles security audit logging (GAP-SEC-008)
type AuditService struct {
	db *gorm.DB
}

// NewAuditService creates a new audit service
func NewAuditService() *AuditService {
	return &AuditService{
		db: database.GetDB(),
	}
}

// AuditEventDetails represents structured details for audit events
type AuditEventDetails struct {
	Email      string                 `json:"email,omitempty"`
	Reason     string                 `json:"reason,omitempty"`
	Changes    map[string]interface{} `json:"changes,omitempty"`
	OldValue   interface{}            `json:"old_value,omitempty"`
	NewValue   interface{}            `json:"new_value,omitempty"`
	Additional map[string]interface{} `json:"additional,omitempty"`
}

// LogEvent creates an audit log entry
func (s *AuditService) LogEvent(
	userID *uuid.UUID,
	action string,
	resource string,
	resourceID *uuid.UUID,
	status string,
	ipAddress string,
	userAgent string,
	details *AuditEventDetails,
) error {
	var detailsJSON string
	if details != nil {
		bytes, err := json.Marshal(details)
		if err != nil {
			detailsJSON = "{}"
		} else {
			detailsJSON = string(bytes)
		}
	}

	auditLog := models.AuditLog{
		UserID:     userID,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		Status:     status,
		Details:    detailsJSON,
		Timestamp:  time.Now(),
	}

	return s.db.Create(&auditLog).Error
}

// LogLogin logs login attempts
func (s *AuditService) LogLogin(userID *uuid.UUID, email string, success bool, ipAddress string, userAgent string, reason string) error {
	status := "success"
	if !success {
		status = "failure"
	}

	return s.LogEvent(
		userID,
		"login",
		"user",
		userID,
		status,
		ipAddress,
		userAgent,
		&AuditEventDetails{
			Email:  email,
			Reason: reason,
		},
	)
}

// LogLogout logs logout events
func (s *AuditService) LogLogout(userID uuid.UUID, ipAddress string, userAgent string) error {
	return s.LogEvent(
		&userID,
		"logout",
		"user",
		&userID,
		"success",
		ipAddress,
		userAgent,
		nil,
	)
}

// LogPasswordChange logs password change events
func (s *AuditService) LogPasswordChange(userID uuid.UUID, ipAddress string, userAgent string) error {
	return s.LogEvent(
		&userID,
		"password_change",
		"user",
		&userID,
		"success",
		ipAddress,
		userAgent,
		nil,
	)
}

// LogDataAccess logs access to sensitive data
func (s *AuditService) LogDataAccess(userID uuid.UUID, resource string, resourceID uuid.UUID, ipAddress string, userAgent string) error {
	return s.LogEvent(
		&userID,
		"data_access",
		resource,
		&resourceID,
		"success",
		ipAddress,
		userAgent,
		nil,
	)
}

// LogDataModification logs modifications to data
func (s *AuditService) LogDataModification(
	userID uuid.UUID,
	action string, // create, update, delete
	resource string,
	resourceID uuid.UUID,
	changes map[string]interface{},
	ipAddress string,
	userAgent string,
) error {
	return s.LogEvent(
		&userID,
		action,
		resource,
		&resourceID,
		"success",
		ipAddress,
		userAgent,
		&AuditEventDetails{
			Changes: changes,
		},
	)
}

// LogAPIKeyCreated logs API key creation
func (s *AuditService) LogAPIKeyCreated(userID uuid.UUID, keyID uuid.UUID, ipAddress string, userAgent string) error {
	return s.LogEvent(
		&userID,
		"api_key_created",
		"api_key",
		&keyID,
		"success",
		ipAddress,
		userAgent,
		nil,
	)
}

// LogAPIKeyDeleted logs API key deletion
func (s *AuditService) LogAPIKeyDeleted(userID uuid.UUID, keyID uuid.UUID, ipAddress string, userAgent string) error {
	return s.LogEvent(
		&userID,
		"api_key_deleted",
		"api_key",
		&keyID,
		"success",
		ipAddress,
		userAgent,
		nil,
	)
}

// LogAccountLocked logs account lockout events
func (s *AuditService) LogAccountLocked(userID uuid.UUID, email string, reason string, ipAddress string, userAgent string) error {
	return s.LogEvent(
		&userID,
		"account_locked",
		"user",
		&userID,
		"success",
		ipAddress,
		userAgent,
		&AuditEventDetails{
			Email:  email,
			Reason: reason,
		},
	)
}

// LogAccountUnlocked logs account unlock events
func (s *AuditService) LogAccountUnlocked(userID uuid.UUID, email string, ipAddress string, userAgent string) error {
	return s.LogEvent(
		&userID,
		"account_unlocked",
		"user",
		&userID,
		"success",
		ipAddress,
		userAgent,
		&AuditEventDetails{
			Email: email,
		},
	)
}

// GetAuditLogs retrieves audit logs with filtering
func (s *AuditService) GetAuditLogs(
	userID *uuid.UUID,
	action string,
	resource string,
	startDate time.Time,
	endDate time.Time,
	limit int,
	offset int,
) ([]models.AuditLog, error) {
	query := s.db.Model(&models.AuditLog{})

	if userID != nil {
		query = query.Where("user_id = ?", userID)
	}

	if action != "" {
		query = query.Where("action = ?", action)
	}

	if resource != "" {
		query = query.Where("resource = ?", resource)
	}

	if !startDate.IsZero() {
		query = query.Where("timestamp >= ?", startDate)
	}

	if !endDate.IsZero() {
		query = query.Where("timestamp <= ?", endDate)
	}

	if limit <= 0 {
		limit = 100
	}

	var logs []models.AuditLog
	err := query.
		Order("timestamp DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error

	return logs, err
}

// CleanupOldLogs removes audit logs older than retention period (1 year for SOC 2)
func (s *AuditService) CleanupOldLogs(retentionDays int) error {
	if retentionDays <= 0 {
		retentionDays = 365 // Default 1 year retention
	}

	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)
	return s.db.Where("created_at < ?", cutoffDate).Delete(&models.AuditLog{}).Error
}
