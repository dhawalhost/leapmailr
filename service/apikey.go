package service

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/dhawalhost/leapmailr/database"
	"github.com/dhawalhost/leapmailr/models"
	"github.com/google/uuid"
)

// GenerateAPIKeyPair creates a new public/private key pair for a user
func GenerateAPIKeyPair(req models.CreateAPIKeyPairRequest, userID uuid.UUID) (*models.APIKeyPairResponse, error) {
	// Generate public and private keys
	publicKey, err := generateKey("pk_live_") // pk = public key
	if err != nil {
		return nil, fmt.Errorf("failed to generate public key: %w", err)
	}

	privateKey, err := generateKey("sk_live_") // sk = secret key
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	// Marshal permissions to JSON
	var permissionsJSON string
	if len(req.Permissions) > 0 {
		permsBytes, err := json.Marshal(req.Permissions)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal permissions: %w", err)
		}
		permissionsJSON = string(permsBytes)
	} else {
		permissionsJSON = "[]"
	}

	// Set default rate limit if not provided
	rateLimit := req.RateLimit
	if rateLimit == 0 {
		rateLimit = 100 // Default: 100 requests per minute
	}

	keyPair := models.APIKeyPair{
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		PublicKey:   publicKey,
		PrivateKey:  privateKey,
		Permissions: permissionsJSON,
		IsActive:    true,
		RateLimit:   rateLimit,
		ExpiresAt:   req.ExpiresAt,
	}

	if err := database.DB.Create(&keyPair).Error; err != nil {
		return nil, fmt.Errorf("failed to create API key pair: %w", err)
	}

	return toAPIKeyPairResponse(&keyPair, true), nil
}

// ListAPIKeyPairs retrieves all API key pairs for a user
func ListAPIKeyPairs(userID uuid.UUID) ([]models.APIKeyPairResponse, error) {
	var keyPairs []models.APIKeyPair
	err := database.DB.Where("user_id = ?", userID).Order("created_at DESC").Find(&keyPairs).Error
	if err != nil {
		return nil, err
	}

	responses := make([]models.APIKeyPairResponse, len(keyPairs))
	for i, kp := range keyPairs {
		responses[i] = *toAPIKeyPairResponse(&kp, false) // Don't include private key in list
	}

	return responses, nil
}

// GetAPIKeyPair retrieves a single API key pair by ID
func GetAPIKeyPair(keyID uuid.UUID, userID uuid.UUID) (*models.APIKeyPairResponse, error) {
	var keyPair models.APIKeyPair
	err := database.DB.Where("id = ? AND user_id = ?", keyID, userID).First(&keyPair).Error
	if err != nil {
		return nil, err
	}

	return toAPIKeyPairResponse(&keyPair, false), nil
}

// UpdateAPIKeyPair updates an existing API key pair
func UpdateAPIKeyPair(keyID uuid.UUID, userID uuid.UUID, req models.UpdateAPIKeyPairRequest) error {
	updates := make(map[string]interface{})

	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if len(req.Permissions) > 0 {
		permsBytes, err := json.Marshal(req.Permissions)
		if err != nil {
			return fmt.Errorf("failed to marshal permissions: %w", err)
		}
		updates["permissions"] = string(permsBytes)
	}
	if req.RateLimit > 0 {
		updates["rate_limit"] = req.RateLimit
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	result := database.DB.Model(&models.APIKeyPair{}).
		Where("id = ? AND user_id = ?", keyID, userID).
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("API key not found")
	}

	return nil
}

// RevokeAPIKeyPair revokes an API key pair
func RevokeAPIKeyPair(keyID uuid.UUID, userID uuid.UUID) error {
	now := time.Now()
	result := database.DB.Model(&models.APIKeyPair{}).
		Where("id = ? AND user_id = ?", keyID, userID).
		Updates(map[string]interface{}{
			"is_active":  false,
			"revoked_at": now,
		})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("API key not found")
	}

	return nil
}

// DeleteAPIKeyPair permanently deletes an API key pair
func DeleteAPIKeyPair(keyID uuid.UUID, userID uuid.UUID) error {
	result := database.DB.Where("id = ? AND user_id = ?", keyID, userID).Delete(&models.APIKeyPair{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("API key not found")
	}

	return nil
}

// RotateAPIKeyPair generates a new private key for an existing key pair
func RotateAPIKeyPair(keyID uuid.UUID, userID uuid.UUID) (*models.APIKeyPairResponse, error) {
	// Get existing key pair
	var keyPair models.APIKeyPair
	err := database.DB.Where("id = ? AND user_id = ?", keyID, userID).First(&keyPair).Error
	if err != nil {
		return nil, err
	}

	// Generate new private key
	newPrivateKey, err := generateKey("sk_live_")
	if err != nil {
		return nil, fmt.Errorf("failed to generate new private key: %w", err)
	}

	// Update the key pair
	keyPair.PrivateKey = newPrivateKey
	if err := database.DB.Save(&keyPair).Error; err != nil {
		return nil, fmt.Errorf("failed to update API key pair: %w", err)
	}

	return toAPIKeyPairResponse(&keyPair, true), nil
}

// ValidateAPIKeyPair validates a public/private key pair
func ValidateAPIKeyPair(publicKey, privateKey string) (*models.APIKeyPair, error) {
	var keyPair models.APIKeyPair
	err := database.DB.Where("public_key = ? AND private_key = ? AND is_active = ?", publicKey, privateKey, true).First(&keyPair).Error
	if err != nil {
		return nil, errors.New("invalid or inactive API key")
	}

	// Check if expired
	if keyPair.ExpiresAt != nil && keyPair.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("API key has expired")
	}

	// Update last used timestamp and usage count
	now := time.Now()
	database.DB.Model(&keyPair).Updates(map[string]interface{}{
		"last_used_at": now,
		"usage_count":  keyPair.UsageCount + 1,
	})

	return &keyPair, nil
}

// ValidatePublicKey validates just the public key (for public operations)
func ValidatePublicKey(publicKey string) (*models.APIKeyPair, error) {
	var keyPair models.APIKeyPair
	err := database.DB.Where("public_key = ? AND is_active = ?", publicKey, true).First(&keyPair).Error
	if err != nil {
		return nil, errors.New("invalid or inactive API key")
	}

	// Check if expired
	if keyPair.ExpiresAt != nil && keyPair.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("API key has expired")
	}

	return &keyPair, nil
}

// LogAPIKeyUsage logs API key usage
func LogAPIKeyUsage(keyID uuid.UUID, endpoint, method string, statusCode int, ipAddress, userAgent string, responseMs int64) error {
	log := models.APIKeyUsageLog{
		APIKeyID:   keyID,
		Endpoint:   endpoint,
		Method:     method,
		StatusCode: statusCode,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		ResponseMs: responseMs,
	}

	return database.DB.Create(&log).Error
}

// GetAPIKeyUsageStats retrieves usage statistics for an API key
func GetAPIKeyUsageStats(keyID uuid.UUID, userID uuid.UUID, limit int) ([]models.APIKeyUsageLog, error) {
	// Verify the key belongs to the user
	var keyPair models.APIKeyPair
	if err := database.DB.Where("id = ? AND user_id = ?", keyID, userID).First(&keyPair).Error; err != nil {
		return nil, errors.New("API key not found")
	}

	var logs []models.APIKeyUsageLog
	query := database.DB.Where("api_key_id = ?", keyID).Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&logs).Error
	return logs, err
}

// Helper functions

func generateKey(prefix string) (string, error) {
	// Generate 32 random bytes
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	// Encode to base64 and add prefix
	encoded := base64.URLEncoding.EncodeToString(b)
	return prefix + encoded[:40], nil // Limit to 40 chars after prefix
}

func toAPIKeyPairResponse(kp *models.APIKeyPair, includePrivateKey bool) *models.APIKeyPairResponse {
	var permissions []string
	if kp.Permissions != "" {
		_ = json.Unmarshal([]byte(kp.Permissions), &permissions)
	}

	response := &models.APIKeyPairResponse{
		ID:          kp.ID,
		Name:        kp.Name,
		Description: kp.Description,
		PublicKey:   kp.PublicKey,
		Permissions: permissions,
		IsActive:    kp.IsActive,
		RateLimit:   kp.RateLimit,
		UsageCount:  kp.UsageCount,
		LastUsedAt:  kp.LastUsedAt,
		ExpiresAt:   kp.ExpiresAt,
		CreatedAt:   kp.CreatedAt,
		UpdatedAt:   kp.UpdatedAt,
		RevokedAt:   kp.RevokedAt,
	}

	if includePrivateKey {
		response.PrivateKey = kp.PrivateKey
	}

	return response
}
