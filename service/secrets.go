package service

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/dhawalhost/leapmailr/logging"
	"github.com/dhawalhost/leapmailr/utils"
	"go.uber.org/zap"
)

// SecretProvider defines the interface for secret management backends
type SecretProvider interface {
	GetSecret(ctx context.Context, key string) (string, error)
	SetSecret(ctx context.Context, key, value string) error
	DeleteSecret(ctx context.Context, key string) error
	ListSecrets(ctx context.Context) ([]string, error)
	RotateSecret(ctx context.Context, key string) (string, error)
}

// SecretMetadata stores metadata about a secret
type SecretMetadata struct {
	Key          string    `json:"key"`
	CreatedAt    time.Time `json:"created_at"`
	LastRotated  time.Time `json:"last_rotated"`
	RotationDays int       `json:"rotation_days"`
	Version      int       `json:"version"`
}

// SecretsManager manages secrets with rotation and encryption
type SecretsManager struct {
	provider     SecretProvider
	logger       *zap.Logger
	encryption   *utils.EncryptionService
	metadata     map[string]*SecretMetadata
	mu           sync.RWMutex
	metadataFile string
}

// NewSecretsManager creates a new secrets manager
func NewSecretsManager(provider SecretProvider, logger *zap.Logger, encryption *utils.EncryptionService) (*SecretsManager, error) {
	if encryption == nil {
		return nil, errors.New("encryption service cannot be nil")
	}

	sm := &SecretsManager{
		provider:     provider,
		logger:       logger,
		encryption:   encryption,
		metadata:     make(map[string]*SecretMetadata),
		metadataFile: "./secrets/metadata.json",
	}

	// Load metadata
	if err := sm.loadMetadata(); err != nil {
		logger.Warn("Failed to load metadata, starting fresh", zap.Error(err))
	}

	return sm, nil
}

// GetSecret retrieves and decrypts a secret
func (sm *SecretsManager) GetSecret(ctx context.Context, key string) (string, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	encrypted, err := sm.provider.GetSecret(ctx, key)
	if err != nil {
		logging.ErrorLog(ctx, err, "Failed to get secret", map[string]interface{}{
			"key": key,
		})
		return "", err
	}

	decrypted, err := sm.encryption.Decrypt(encrypted)
	if err != nil {
		logging.ErrorLog(ctx, err, "Failed to decrypt secret", map[string]interface{}{
			"key": key,
		})
		return "", err
	}

	logging.AuditLog(ctx, "secret_access", "secret", map[string]interface{}{
		"key":    key,
		"action": "read",
	})

	return decrypted, nil
}

// SetSecret encrypts and stores a secret
func (sm *SecretsManager) SetSecret(ctx context.Context, key, value string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	encrypted, err := sm.encryption.Encrypt(value)
	if err != nil {
		logging.ErrorLog(ctx, err, "Failed to encrypt secret", map[string]interface{}{
			"key": key,
		})
		return err
	}

	if err := sm.provider.SetSecret(ctx, key, encrypted); err != nil {
		logging.ErrorLog(ctx, err, "Failed to store secret", map[string]interface{}{
			"key": key,
		})
		return err
	}

	// Update metadata
	now := time.Now()
	if meta, exists := sm.metadata[key]; exists {
		meta.LastRotated = now
		meta.Version++
	} else {
		sm.metadata[key] = &SecretMetadata{
			Key:          key,
			CreatedAt:    now,
			LastRotated:  now,
			RotationDays: 90, // Default 90 days
			Version:      1,
		}
	}

	if err := sm.saveMetadata(); err != nil {
		sm.logger.Warn("Failed to save metadata", zap.Error(err))
	}

	logging.AuditLog(ctx, "secret_update", "secret", map[string]interface{}{
		"key":    key,
		"action": "write",
	})

	return nil
}

// DeleteSecret removes a secret
func (sm *SecretsManager) DeleteSecret(ctx context.Context, key string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if err := sm.provider.DeleteSecret(ctx, key); err != nil {
		logging.ErrorLog(ctx, err, "Failed to delete secret", map[string]interface{}{
			"key": key,
		})
		return err
	}

	delete(sm.metadata, key)
	if err := sm.saveMetadata(); err != nil {
		sm.logger.Warn("Failed to save metadata", zap.Error(err))
	}

	logging.AuditLog(ctx, "secret_delete", "secret", map[string]interface{}{
		"key":    key,
		"action": "delete",
	})

	return nil
}

// RotateSecret generates a new value for a secret
func (sm *SecretsManager) RotateSecret(ctx context.Context, key string) (string, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	newValue, err := sm.provider.RotateSecret(ctx, key)
	if err != nil {
		logging.ErrorLog(ctx, err, "Failed to rotate secret", map[string]interface{}{
			"key": key,
		})
		return "", err
	}

	encrypted, err := sm.encryption.Encrypt(newValue)
	if err != nil {
		logging.ErrorLog(ctx, err, "Failed to encrypt rotated secret", map[string]interface{}{
			"key": key,
		})
		return "", err
	}

	if err := sm.provider.SetSecret(ctx, key, encrypted); err != nil {
		logging.ErrorLog(ctx, err, "Failed to store rotated secret", map[string]interface{}{
			"key": key,
		})
		return "", err
	}

	// Update metadata
	now := time.Now()
	if meta, exists := sm.metadata[key]; exists {
		meta.LastRotated = now
		meta.Version++
	}

	if err := sm.saveMetadata(); err != nil {
		sm.logger.Warn("Failed to save metadata", zap.Error(err))
	}

	logging.AuditLog(ctx, "secret_rotate", "secret", map[string]interface{}{
		"key":    key,
		"action": "rotate",
	})

	sm.logger.Info("Secret rotated successfully",
		zap.String("key", key),
		zap.Time("rotated_at", now))

	return newValue, nil
}

// CheckRotationNeeded checks if secrets need rotation
func (sm *SecretsManager) CheckRotationNeeded() []string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var needRotation []string
	now := time.Now()

	for key, meta := range sm.metadata {
		daysSinceRotation := int(now.Sub(meta.LastRotated).Hours() / 24)
		if daysSinceRotation >= meta.RotationDays {
			needRotation = append(needRotation, key)
		}
	}

	return needRotation
}

// GetMetadata returns metadata for a secret
func (sm *SecretsManager) GetMetadata(key string) (*SecretMetadata, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	meta, exists := sm.metadata[key]
	if !exists {
		return nil, errors.New("metadata not found")
	}

	return meta, nil
}

// ListSecrets returns all secret keys
func (sm *SecretsManager) ListSecrets(ctx context.Context) ([]string, error) {
	return sm.provider.ListSecrets(ctx)
}

// loadMetadata loads secret metadata from disk
func (sm *SecretsManager) loadMetadata() error {
	data, err := ioutil.ReadFile(sm.metadataFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist yet, that's ok
		}
		return err
	}

	return json.Unmarshal(data, &sm.metadata)
}

// saveMetadata saves secret metadata to disk
func (sm *SecretsManager) saveMetadata() error {
	// Ensure directory exists
	if err := os.MkdirAll("./secrets", 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(sm.metadata, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(sm.metadataFile, data, 0600)
}

// StartRotationMonitor starts a background goroutine to monitor rotation needs
func (sm *SecretsManager) StartRotationMonitor(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	sm.logger.Info("Started secret rotation monitor",
		zap.Duration("interval", interval))

	for {
		select {
		case <-ctx.Done():
			sm.logger.Info("Stopping secret rotation monitor")
			return
		case <-ticker.C:
			needRotation := sm.CheckRotationNeeded()
			if len(needRotation) > 0 {
				sm.logger.Warn("Secrets need rotation",
					zap.Strings("keys", needRotation),
					zap.Int("count", len(needRotation)))

				// Alert via logging
				logging.SecurityLog(ctx, "secrets_rotation_needed", map[string]interface{}{
					"keys":  needRotation,
					"count": len(needRotation),
				})
			}
		}
	}
}
