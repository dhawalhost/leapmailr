package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dhawalhost/leapmailr/database"
	"github.com/dhawalhost/leapmailr/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SuppressionService struct {
	db *gorm.DB
}

func NewSuppressionService() *SuppressionService {
	return &SuppressionService{
		db: database.GetDB(),
	}
}

// IsEmailSuppressed checks if an email address is in the suppression list
func (s *SuppressionService) IsEmailSuppressed(email string, userID uuid.UUID) (bool, *models.Suppression, error) {
	var suppression models.Suppression
	email = strings.ToLower(strings.TrimSpace(email))

	err := s.db.Where("email = ? AND (user_id = ? OR user_id IS NULL)", email, userID).
		First(&suppression).Error

	if err == gorm.ErrRecordNotFound {
		return false, nil, nil
	}
	if err != nil {
		return false, nil, err
	}

	return true, &suppression, nil
}

// AddSuppression adds an email to the suppression list
func (s *SuppressionService) AddSuppression(req models.CreateSuppressionRequest, userID uuid.UUID) (*models.SuppressionResponse, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))

	// Check if already suppressed
	var existing models.Suppression
	if err := s.db.Where("email = ? AND user_id = ?", email, userID).First(&existing).Error; err == nil {
		return nil, fmt.Errorf("email already in suppression list")
	}

	// Marshal metadata
	var metadataJSON string
	if req.Metadata != nil {
		metadataBytes, err := json.Marshal(req.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}
		metadataJSON = string(metadataBytes)
	}

	suppression := models.Suppression{
		UserID:   &userID,
		Email:    email,
		Reason:   req.Reason,
		Source:   models.SourceManual,
		Metadata: metadataJSON,
	}

	if err := s.db.Create(&suppression).Error; err != nil {
		return nil, err
	}

	return &models.SuppressionResponse{
		ID:        suppression.ID,
		Email:     suppression.Email,
		Reason:    suppression.Reason,
		Source:    suppression.Source,
		Metadata:  suppression.Metadata,
		CreatedAt: suppression.CreatedAt,
	}, nil
}

// AddBulkSuppressions adds multiple emails to the suppression list
func (s *SuppressionService) AddBulkSuppressions(req models.BulkSuppressionRequest, userID uuid.UUID) (int, error) {
	var metadataJSON string
	if req.Metadata != nil {
		metadataBytes, err := json.Marshal(req.Metadata)
		if err != nil {
			return 0, fmt.Errorf("failed to marshal metadata: %w", err)
		}
		metadataJSON = string(metadataBytes)
	}

	added := 0
	for _, email := range req.Emails {
		email = strings.ToLower(strings.TrimSpace(email))

		// Check if already exists
		var existing models.Suppression
		if err := s.db.Where("email = ? AND user_id = ?", email, userID).First(&existing).Error; err == nil {
			continue // Skip if already exists
		}

		suppression := models.Suppression{
			UserID:   &userID,
			Email:    email,
			Reason:   req.Reason,
			Source:   models.SourceAPI,
			Metadata: metadataJSON,
		}

		if err := s.db.Create(&suppression).Error; err == nil {
			added++
		}
	}

	return added, nil
}

// ListSuppressions lists suppression entries with optional filters
func (s *SuppressionService) ListSuppressions(userID uuid.UUID, filters models.SuppressionFilters) ([]models.SuppressionResponse, int64, error) {
	query := s.db.Model(&models.Suppression{}).Where("user_id = ?", userID)

	// Apply filters
	if filters.Reason != "" {
		query = query.Where("reason = ?", filters.Reason)
	}
	if filters.Source != "" {
		query = query.Where("source = ?", filters.Source)
	}
	if filters.Search != "" {
		query = query.Where("email LIKE ?", "%"+filters.Search+"%")
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	} else {
		query = query.Limit(50) // Default limit
	}
	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}

	// Order by most recent
	query = query.Order("created_at DESC")

	var suppressions []models.Suppression
	if err := query.Find(&suppressions).Error; err != nil {
		return nil, 0, err
	}

	var response []models.SuppressionResponse
	for _, s := range suppressions {
		response = append(response, models.SuppressionResponse{
			ID:        s.ID,
			Email:     s.Email,
			Reason:    s.Reason,
			Source:    s.Source,
			Metadata:  s.Metadata,
			CreatedAt: s.CreatedAt,
		})
	}

	return response, total, nil
}

// DeleteSuppression removes an email from the suppression list
func (s *SuppressionService) DeleteSuppression(suppressionID, userID uuid.UUID) error {
	result := s.db.Where("id = ? AND user_id = ?", suppressionID, userID).Delete(&models.Suppression{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("suppression not found")
	}
	return nil
}

// AddSuppressionFromWebhook adds a suppression from an email provider webhook
func (s *SuppressionService) AddSuppressionFromWebhook(email string, reason models.SuppressionReason, metadata map[string]interface{}, userID *uuid.UUID) error {
	email = strings.ToLower(strings.TrimSpace(email))

	// Check if already exists
	var existing models.Suppression
	query := s.db.Where("email = ?", email)
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}

	if err := query.First(&existing).Error; err == nil {
		// Already exists, update metadata if needed
		if metadata != nil {
			metadataBytes, _ := json.Marshal(metadata)
			existing.Metadata = string(metadataBytes)
			s.db.Save(&existing)
		}
		return nil
	}

	// Marshal metadata
	var metadataJSON string
	if metadata != nil {
		metadataBytes, err := json.Marshal(metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
		metadataJSON = string(metadataBytes)
	}

	suppression := models.Suppression{
		UserID:   userID,
		Email:    email,
		Reason:   reason,
		Source:   models.SourceWebhook,
		Metadata: metadataJSON,
	}

	return s.db.Create(&suppression).Error
}
