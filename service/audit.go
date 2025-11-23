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

// AuditEventParams holds parameters for audit logging
type AuditEventParams struct {
	UserID     *uuid.UUID
	Action     string
	Resource   string
	ResourceID *uuid.UUID
	Status     string
	IPAddress  string
	UserAgent  string
	Details    *AuditEventDetails
}

// LogEvent creates an audit log entry
func (s *AuditService) LogEvent(params *AuditEventParams) error {
	var detailsJSON string
	if params.Details != nil {
		bytes, err := json.Marshal(params.Details)
		if err != nil {
			detailsJSON = "{}"
		} else {
			detailsJSON = string(bytes)
		}
	}

	auditLog := models.AuditLog{
		UserID:     params.UserID,
		Action:     params.Action,
		Resource:   params.Resource,
		ResourceID: params.ResourceID,
		IPAddress:  params.IPAddress,
		UserAgent:  params.UserAgent,
		Status:     params.Status,
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

	return s.LogEvent(&AuditEventParams{
		UserID:     userID,
		Action:     "login",
		Resource:   "user",
		ResourceID: userID,
		Status:     status,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		Details: &AuditEventDetails{
			Email:  email,
			Reason: reason,
		},
	})
}

// LogLogout logs logout events
func (s *AuditService) LogLogout(userID uuid.UUID, ipAddress string, userAgent string) error {
	return s.LogEvent(&AuditEventParams{
		UserID:     &userID,
		Action:     "logout",
		Resource:   "user",
		ResourceID: &userID,
		Status:     "success",
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
	})
}

// LogPasswordChange logs password change events
func (s *AuditService) LogPasswordChange(userID uuid.UUID, ipAddress string, userAgent string) error {
	return s.LogEvent(&AuditEventParams{
		UserID:     &userID,
		Action:     "password_change",
		Resource:   "user",
		ResourceID: &userID,
		Status:     "success",
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
	})
}

// LogDataAccess logs access to sensitive data
func (s *AuditService) LogDataAccess(userID uuid.UUID, resource string, resourceID uuid.UUID, ipAddress string, userAgent string) error {
	return s.LogEvent(&AuditEventParams{
		UserID:     &userID,
		Action:     "data_access",
		Resource:   resource,
		ResourceID: &resourceID,
		Status:     "success",
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
	})
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
	return s.LogEvent(&AuditEventParams{
		UserID:     &userID,
		Action:     action,
		Resource:   resource,
		ResourceID: &resourceID,
		Status:     "success",
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		Details: &AuditEventDetails{
			Changes: changes,
		},
	})
}

// LogAPIKeyCreated logs API key creation
func (s *AuditService) LogAPIKeyCreated(userID uuid.UUID, keyID uuid.UUID, ipAddress string, userAgent string) error {
	return s.LogEvent(&AuditEventParams{
		UserID:     &userID,
		Action:     "api_key_created",
		Resource:   "api_key",
		ResourceID: &keyID,
		Status:     "success",
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
	})
}

// LogAPIKeyDeleted logs API key deletion
func (s *AuditService) LogAPIKeyDeleted(userID uuid.UUID, keyID uuid.UUID, ipAddress string, userAgent string) error {
	return s.LogEvent(&AuditEventParams{
		UserID:     &userID,
		Action:     "api_key_deleted",
		Resource:   "api_key",
		ResourceID: &keyID,
		Status:     "success",
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
	})
}

// LogAccountLocked logs account lockout events
func (s *AuditService) LogAccountLocked(userID uuid.UUID, email string, reason string, ipAddress string, userAgent string) error {
	return s.LogEvent(&AuditEventParams{
		UserID:     &userID,
		Action:     "account_locked",
		Resource:   "user",
		ResourceID: &userID,
		Status:     "success",
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		Details: &AuditEventDetails{
			Email:  email,
			Reason: reason,
		},
	})
}

// LogAccountUnlocked logs account unlock events
func (s *AuditService) LogAccountUnlocked(userID uuid.UUID, email string, ipAddress string, userAgent string) error {
	return s.LogEvent(&AuditEventParams{
		UserID:     &userID,
		Action:     "account_unlocked",
		Resource:   "user",
		ResourceID: &userID,
		Status:     "success",
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		Details: &AuditEventDetails{
			Email: email,
		},
	})
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
