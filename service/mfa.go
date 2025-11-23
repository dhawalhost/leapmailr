package service

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image/png"
	"time"

	"github.com/dhawalhost/leapmailr/database"
	"github.com/dhawalhost/leapmailr/models"
	"github.com/dhawalhost/leapmailr/monitoring"
	"github.com/dhawalhost/leapmailr/utils"
	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
)

const (
	backupCodeLength = 8
	backupCodeCount  = 10
	issuer           = "LeapMailR"
	queryIDEquals    = "id = ?"
)

var (
	ErrMFANotEnabled     = errors.New("MFA is not enabled for this user")
	ErrMFAAlreadyEnabled = errors.New("MFA is already enabled for this user")
	ErrInvalidMFACode    = errors.New("invalid MFA code")
	ErrInvalidBackupCode = errors.New("invalid backup code")
	ErrNoBackupCodesLeft = errors.New("no backup codes remaining")
	ErrInvalidPassword   = errors.New("invalid password")
	ErrMFANotVerified    = errors.New("MFA setup not completed")
)

// MFAService handles multi-factor authentication operations
type MFAService struct {
	encryption *utils.EncryptionService
	audit      *AuditService
}

// NewMFAService creates a new MFA service
func NewMFAService(encryption *utils.EncryptionService) *MFAService {
	return &MFAService{
		encryption: encryption,
		audit:      NewAuditService(),
	}
}

// SetupMFA initiates MFA setup for a user
func (s *MFAService) SetupMFA(userID string, password string) (*models.MFASetupResponse, error) {
	db := database.GetDB()

	// Get user
	var user models.User
	if err := db.Where(queryIDEquals, userID).First(&user).Error; err != nil {
		monitoring.ErrorsTotal.WithLabelValues("user_not_found", "mfa").Inc()
		return nil, err
	}

	// Check if MFA is already enabled
	if user.MFAEnabled {
		return nil, ErrMFAAlreadyEnabled
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		monitoring.AuthFailuresTotal.WithLabelValues("mfa_setup", "invalid_password").Inc()
		return nil, ErrInvalidPassword
	}

	// Generate TOTP secret
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: user.Email,
		SecretSize:  32,
	})
	if err != nil {
		monitoring.ErrorsTotal.WithLabelValues("totp_generation_failed", "mfa").Inc()
		return nil, err
	}

	// Generate QR code
	var buf bytes.Buffer
	img, err := key.Image(256, 256)
	if err != nil {
		monitoring.ErrorsTotal.WithLabelValues("qr_generation_failed", "mfa").Inc()
		return nil, err
	}

	if err := png.Encode(&buf, img); err != nil {
		monitoring.ErrorsTotal.WithLabelValues("qr_encoding_failed", "mfa").Inc()
		return nil, err
	}

	qrCodeDataURL := fmt.Sprintf("data:image/png;base64,%s", base64.StdEncoding.EncodeToString(buf.Bytes()))

	// Generate backup codes
	backupCodes, err := s.generateBackupCodes()
	if err != nil {
		monitoring.ErrorsTotal.WithLabelValues("backup_code_generation_failed", "mfa").Inc()
		return nil, err
	}

	// Hash and encrypt backup codes for storage
	hashedCodes, err := s.hashAndEncryptBackupCodes(backupCodes)
	if err != nil {
		monitoring.ErrorsTotal.WithLabelValues("backup_code_encryption_failed", "mfa").Inc()
		return nil, err
	}

	// Encrypt TOTP secret
	encryptedSecret, err := s.encryption.Encrypt(key.Secret())
	if err != nil {
		monitoring.ErrorsTotal.WithLabelValues("secret_encryption_failed", "mfa").Inc()
		return nil, err
	}

	// Store encrypted secret and backup codes (not yet enabled)
	user.MFASecret = encryptedSecret
	user.MFABackupCodes = hashedCodes

	if err := db.Save(&user).Error; err != nil {
		monitoring.ErrorsTotal.WithLabelValues("database_error", "mfa").Inc()
		return nil, err
	}

	// Log MFA setup initiated
	userUUID, _ := uuid.Parse(userID)
	s.audit.LogEvent(&userUUID, "mfa_setup_initiated", "user", &userUUID, "success", "", "", nil)

	return &models.MFASetupResponse{
		Secret:        key.Secret(),
		QRCodeDataURL: qrCodeDataURL,
		BackupCodes:   backupCodes,
	}, nil
}

// VerifyMFASetup verifies the TOTP code and enables MFA
func (s *MFAService) VerifyMFASetup(userID string, code string) error {
	db := database.GetDB()

	var user models.User
	if err := db.Where(queryIDEquals, userID).First(&user).Error; err != nil {
		monitoring.ErrorsTotal.WithLabelValues("user_not_found", "mfa").Inc()
		return err
	}

	// Check if MFA is already enabled
	if user.MFAEnabled {
		return ErrMFAAlreadyEnabled
	}

	// Check if secret exists
	if user.MFASecret == "" {
		return errors.New("MFA setup not initiated")
	}

	// Decrypt secret
	secret, err := s.encryption.Decrypt(user.MFASecret)
	if err != nil {
		monitoring.ErrorsTotal.WithLabelValues("secret_decryption_failed", "mfa").Inc()
		return err
	}

	// Verify TOTP code
	valid := totp.Validate(code, secret)
	if !valid {
		monitoring.AuthFailuresTotal.WithLabelValues("mfa_verify", "invalid_code").Inc()
		return ErrInvalidMFACode
	}

	// Enable MFA
	now := time.Now()
	user.MFAEnabled = true
	user.MFAVerifiedAt = &now

	if err := db.Save(&user).Error; err != nil {
		monitoring.ErrorsTotal.WithLabelValues("database_error", "mfa").Inc()
		return err
	}

	// Log MFA enabled
	userUUID, _ := uuid.Parse(userID)
	s.audit.LogEvent(&userUUID, "mfa_enabled", "user", &userUUID, "success", "", "", nil)

	return nil
}

// VerifyMFACode verifies a TOTP code during login
func (s *MFAService) VerifyMFACode(userID string, code string) (bool, error) {
	db := database.GetDB()

	var user models.User
	if err := db.Where(queryIDEquals, userID).First(&user).Error; err != nil {
		monitoring.ErrorsTotal.WithLabelValues("user_not_found", "mfa").Inc()
		return false, err
	}

	if !user.MFAEnabled {
		return false, ErrMFANotEnabled
	}

	// Decrypt secret
	secret, err := s.encryption.Decrypt(user.MFASecret)
	if err != nil {
		monitoring.ErrorsTotal.WithLabelValues("secret_decryption_failed", "mfa").Inc()
		return false, err
	}

	// Verify TOTP code
	valid := totp.Validate(code, secret)
	if valid {
		monitoring.AuthAttemptsTotal.WithLabelValues("mfa", "success").Inc()
		userUUID, _ := uuid.Parse(userID)
		s.audit.LogEvent(&userUUID, "mfa_code_verified", "user", &userUUID, "success", "", "", nil)
		return true, nil
	}

	monitoring.AuthFailuresTotal.WithLabelValues("mfa_login", "invalid_code").Inc()
	userUUID, _ := uuid.Parse(userID)
	s.audit.LogEvent(&userUUID, "mfa_code_failed", "user", &userUUID, "failure", "", "", &AuditEventDetails{
		Reason: "invalid code",
	})
	return false, nil
}

// VerifyBackupCode verifies a backup code during login
func (s *MFAService) VerifyBackupCode(userID string, backupCode string) (bool, error) {
	db := database.GetDB()

	var user models.User
	if err := db.Where(queryIDEquals, userID).First(&user).Error; err != nil {
		monitoring.ErrorsTotal.WithLabelValues("user_not_found", "mfa").Inc()
		return false, err
	}

	if !user.MFAEnabled {
		return false, ErrMFANotEnabled
	}

	// Decrypt backup codes
	backupCodes, err := s.decryptBackupCodes(user.MFABackupCodes)
	if err != nil {
		monitoring.ErrorsTotal.WithLabelValues("backup_code_decryption_failed", "mfa").Inc()
		return false, err
	}

	if len(backupCodes) == 0 {
		return false, ErrNoBackupCodesLeft
	}

	// Check each backup code
	validCodeIndex := -1
	for i, hashedCode := range backupCodes {
		if err := bcrypt.CompareHashAndPassword([]byte(hashedCode), []byte(backupCode)); err == nil {
			validCodeIndex = i
			break
		}
	}

	if validCodeIndex == -1 {
		monitoring.AuthFailuresTotal.WithLabelValues("mfa_backup_code", "invalid_code").Inc()
		userUUID, _ := uuid.Parse(userID)
		s.audit.LogEvent(&userUUID, "mfa_backup_code_failed", "user", &userUUID, "failure", "", "", &AuditEventDetails{
			Reason: "invalid backup code",
		})
		return false, nil
	}

	// Remove used backup code
	backupCodes = append(backupCodes[:validCodeIndex], backupCodes[validCodeIndex+1:]...)

	// Encrypt and save remaining codes
	encrypted, err := s.encryptBackupCodes(backupCodes)
	if err != nil {
		monitoring.ErrorsTotal.WithLabelValues("backup_code_encryption_failed", "mfa").Inc()
		return false, err
	}

	user.MFABackupCodes = encrypted
	if err := db.Save(&user).Error; err != nil {
		monitoring.ErrorsTotal.WithLabelValues("database_error", "mfa").Inc()
		return false, err
	}

	monitoring.AuthAttemptsTotal.WithLabelValues("mfa_backup_code", "success").Inc()
	userUUID, _ := uuid.Parse(userID)
	s.audit.LogEvent(&userUUID, "mfa_backup_code_used", "user", &userUUID, "success", "", "", &AuditEventDetails{
		Additional: map[string]interface{}{
			"codes_remaining": len(backupCodes),
		},
	})

	return true, nil
}

// DisableMFA disables MFA for a user
func (s *MFAService) DisableMFA(userID string, password string, code string) error {
	db := database.GetDB()

	var user models.User
	if err := db.Where(queryIDEquals, userID).First(&user).Error; err != nil {
		monitoring.ErrorsTotal.WithLabelValues("user_not_found", "mfa").Inc()
		return err
	}

	if !user.MFAEnabled {
		return ErrMFANotEnabled
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		monitoring.AuthFailuresTotal.WithLabelValues("mfa_disable", "invalid_password").Inc()
		return ErrInvalidPassword
	}

	// Verify MFA code
	valid, err := s.VerifyMFACode(userID, code)
	if err != nil {
		return err
	}
	if !valid {
		return ErrInvalidMFACode
	}

	// Disable MFA
	user.MFAEnabled = false
	user.MFASecret = ""
	user.MFABackupCodes = ""
	user.MFAVerifiedAt = nil

	if err := db.Save(&user).Error; err != nil {
		monitoring.ErrorsTotal.WithLabelValues("database_error", "mfa").Inc()
		return err
	}

	// Log MFA disabled
	userUUID, _ := uuid.Parse(userID)
	s.audit.LogEvent(&userUUID, "mfa_disabled", "user", &userUUID, "success", "", "", nil)

	return nil
}

// RegenerateBackupCodes generates new backup codes
func (s *MFAService) RegenerateBackupCodes(userID string, password string, code string) ([]string, error) {
	db := database.GetDB()

	var user models.User
	if err := db.Where(queryIDEquals, userID).First(&user).Error; err != nil {
		monitoring.ErrorsTotal.WithLabelValues("user_not_found", "mfa").Inc()
		return nil, err
	}

	if !user.MFAEnabled {
		return nil, ErrMFANotEnabled
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		monitoring.AuthFailuresTotal.WithLabelValues("mfa_regenerate_codes", "invalid_password").Inc()
		return nil, ErrInvalidPassword
	}

	// Verify MFA code
	valid, err := s.VerifyMFACode(userID, code)
	if err != nil {
		return nil, err
	}
	if !valid {
		return nil, ErrInvalidMFACode
	}

	// Generate new backup codes
	backupCodes, err := s.generateBackupCodes()
	if err != nil {
		monitoring.ErrorsTotal.WithLabelValues("backup_code_generation_failed", "mfa").Inc()
		return nil, err
	}

	// Hash and encrypt backup codes
	hashedCodes, err := s.hashAndEncryptBackupCodes(backupCodes)
	if err != nil {
		monitoring.ErrorsTotal.WithLabelValues("backup_code_encryption_failed", "mfa").Inc()
		return nil, err
	}

	// Save new backup codes
	user.MFABackupCodes = hashedCodes
	if err := db.Save(&user).Error; err != nil {
		monitoring.ErrorsTotal.WithLabelValues("database_error", "mfa").Inc()
		return nil, err
	}

	// Log backup codes regenerated
	userUUID, _ := uuid.Parse(userID)
	s.audit.LogEvent(&userUUID, "mfa_backup_codes_regenerated", "user", &userUUID, "success", "", "", nil)

	return backupCodes, nil
}

// generateBackupCodes generates random backup codes
func (s *MFAService) generateBackupCodes() ([]string, error) {
	codes := make([]string, backupCodeCount)
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // Removed confusing characters (0, O, 1, I)

	for i := 0; i < backupCodeCount; i++ {
		code := make([]byte, backupCodeLength)
		randomBytes := make([]byte, backupCodeLength)

		if _, err := rand.Read(randomBytes); err != nil {
			return nil, err
		}

		for j := 0; j < backupCodeLength; j++ {
			code[j] = charset[int(randomBytes[j])%len(charset)]
		}

		codes[i] = string(code)
	}

	return codes, nil
}

// hashAndEncryptBackupCodes hashes each backup code and encrypts the list
func (s *MFAService) hashAndEncryptBackupCodes(codes []string) (string, error) {
	hashedCodes := make([]string, len(codes))

	for i, code := range codes {
		hash, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
		if err != nil {
			return "", err
		}
		hashedCodes[i] = string(hash)
	}

	// Convert to JSON
	jsonData, err := json.Marshal(hashedCodes)
	if err != nil {
		return "", err
	}

	// Encrypt
	encrypted, err := s.encryption.Encrypt(string(jsonData))
	if err != nil {
		return "", err
	}

	return encrypted, nil
}

// decryptBackupCodes decrypts and returns the list of hashed backup codes
func (s *MFAService) decryptBackupCodes(encrypted string) ([]string, error) {
	if encrypted == "" {
		return []string{}, nil
	}

	// Decrypt
	decrypted, err := s.encryption.Decrypt(encrypted)
	if err != nil {
		return nil, err
	}

	// Parse JSON
	var codes []string
	if err := json.Unmarshal([]byte(decrypted), &codes); err != nil {
		return nil, err
	}

	return codes, nil
}

// encryptBackupCodes encrypts a list of hashed backup codes
func (s *MFAService) encryptBackupCodes(hashedCodes []string) (string, error) {
	// Convert to JSON
	jsonData, err := json.Marshal(hashedCodes)
	if err != nil {
		return "", err
	}

	// Encrypt
	encrypted, err := s.encryption.Encrypt(string(jsonData))
	if err != nil {
		return "", err
	}

	return encrypted, nil
}

// GetMFAStatus returns the MFA status for a user
func (s *MFAService) GetMFAStatus(userID string) (bool, int, error) {
	db := database.GetDB()

	var user models.User
	if err := db.Where(queryIDEquals, userID).First(&user).Error; err != nil {
		return false, 0, err
	}

	if !user.MFAEnabled {
		return false, 0, nil
	}

	// Get remaining backup codes count
	backupCodes, err := s.decryptBackupCodes(user.MFABackupCodes)
	if err != nil {
		return true, 0, err
	}

	return true, len(backupCodes), nil
}
