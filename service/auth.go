package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/dhawalhost/leapmailr/config"
	"github.com/dhawalhost/leapmailr/database"
	"github.com/dhawalhost/leapmailr/models"
	"github.com/dhawalhost/leapmailr/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AuthService handles authentication operations
type AuthService struct {
	db *gorm.DB
}

// NewAuthService creates a new authentication service
func NewAuthService() *AuthService {
	return &AuthService{
		db: database.GetDB(),
	}
}

// Register creates a new user account
func (s *AuthService) Register(req models.RegisterRequest) (*models.AuthResponse, error) {
	// Validate email format (GAP-SEC-013)
	if err := utils.ValidateEmail(req.Email); err != nil {
		return nil, fmt.Errorf("invalid email: %w", err)
	}

	// Sanitize inputs (GAP-SEC-013)
	req.Email = utils.SanitizeInput(req.Email)
	req.FirstName = utils.SanitizeInput(req.FirstName)
	req.LastName = utils.SanitizeInput(req.LastName)

	// Validate name fields
	if err := utils.ValidateName(req.FirstName); err != nil {
		return nil, fmt.Errorf("invalid first name: %w", err)
	}
	if err := utils.ValidateName(req.LastName); err != nil {
		return nil, fmt.Errorf("invalid last name: %w", err)
	}

	// Validate password strength (GAP-SEC-002)
	if err := utils.ValidatePassword(req.Password, utils.DefaultPasswordPolicy()); err != nil {
		return nil, fmt.Errorf("weak password: %w", err)
	}

	// Check if user already exists
	var existingUser models.User
	if err := s.db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		return nil, fmt.Errorf("user with email %s already exists", req.Email)
	} else if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Generate API keys
	apiKey, err := generateRandomKey(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate API key: %w", err)
	}

	privateKey, err := generateRandomKey(64)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	// Encrypt private key (GAP-SEC-005)
	encryption, err := utils.NewEncryptionService()
	if err != nil {
		return nil, fmt.Errorf("encryption service error: %w", err)
	}
	encryptedPrivateKey, err := encryption.Encrypt(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt private key: %w", err)
	}

	// Create user
	user := models.User{
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		APIKey:       apiKey,
		PrivateKey:   encryptedPrivateKey,
		PlanType:     "free",
		IsActive:     true,
	}

	if err := s.db.Create(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Store password hash in history (GAP-SEC-002)
	passwordHistory := models.PasswordHistory{
		UserID:       user.ID,
		PasswordHash: string(hashedPassword),
	}
	if err := s.db.Create(&passwordHistory).Error; err != nil {
		// Log but don't fail registration
		fmt.Printf("Warning: Failed to create password history for user %s: %v\n", user.Email, err)
	}

	// Create default project for the user
	_, err = CreateProject(user.ID, "Default Project", "Your default project", "#3b82f6", true)
	if err != nil {
		// Log the error but don't fail registration
		fmt.Printf("Warning: Failed to create default project for user %s: %v\n", user.Email, err)
	}

	// Generate tokens
	accessToken, refreshToken, expiresIn, err := s.generateTokenPair(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	return &models.AuthResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
	}, nil
}

// Login authenticates a user with account lockout protection (GAP-SEC-003)
func (s *AuthService) Login(req models.LoginRequest, ipAddress, userAgent string) (*models.AuthResponse, error) {
	var user models.User
	if err := s.db.Where("email = ? AND is_active = ?", req.Email, true).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Log failed login attempt for non-existent user
			auditService := NewAuditService()
			_ = auditService.LogLogin(nil, req.Email, false, ipAddress, userAgent, "user not found")
			return nil, fmt.Errorf("invalid credentials")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Check if account is locked (GAP-SEC-003)
	now := time.Now()
	if user.LockedUntil != nil && user.LockedUntil.After(now) {
		remainingMinutes := int(time.Until(*user.LockedUntil).Minutes())
		auditService := NewAuditService()
		_ = auditService.LogLogin(&user.ID, user.Email, false, ipAddress, userAgent,
			fmt.Sprintf("account locked for %d more minutes", remainingMinutes))
		return nil, fmt.Errorf("account is locked due to too many failed login attempts. Try again in %d minutes", remainingMinutes)
	}

	// Update last login attempt time
	user.LastLoginAttempt = &now

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		// Increment failed login attempts
		user.FailedLoginAttempts++

		// Lock account after 5 failed attempts (GAP-SEC-003)
		const maxFailedAttempts = 5
		const lockoutDurationMinutes = 15

		if user.FailedLoginAttempts >= maxFailedAttempts {
			lockoutUntil := now.Add(time.Duration(lockoutDurationMinutes) * time.Minute)
			user.LockedUntil = &lockoutUntil

			// Save lockout state
			if err := s.db.Save(&user).Error; err != nil {
				fmt.Printf("Warning: Failed to lock account for user %s: %v\n", user.Email, err)
			}

			// Log account lockout event
			auditService := NewAuditService()
			_ = auditService.LogAccountLocked(user.ID, user.Email,
				fmt.Sprintf("locked for %d minutes after %d failed attempts", lockoutDurationMinutes, maxFailedAttempts),
				ipAddress, userAgent)

			return nil, fmt.Errorf("account locked due to too many failed login attempts. Try again in %d minutes", lockoutDurationMinutes)
		}

		// Save failed attempt count
		if err := s.db.Save(&user).Error; err != nil {
			fmt.Printf("Warning: Failed to update failed login count for user %s: %v\n", user.Email, err)
		}

		// Log failed login
		auditService := NewAuditService()
		_ = auditService.LogLogin(&user.ID, user.Email, false, ipAddress, userAgent,
			fmt.Sprintf("invalid password (attempt %d/%d)", user.FailedLoginAttempts, maxFailedAttempts))

		return nil, fmt.Errorf("invalid credentials")
	}

	// Successful login - reset failed attempts and unlock
	user.FailedLoginAttempts = 0
	user.LockedUntil = nil
	user.LastLoginSuccess = &now

	// Save success state
	if err := s.db.Save(&user).Error; err != nil {
		fmt.Printf("Warning: Failed to reset failed login count for user %s: %v\n", user.Email, err)
	}

	// Check if MFA is enabled (GAP-SEC-001)
	if user.MFAEnabled {
		// Don't generate tokens yet - require MFA verification
		// Log successful password verification
		auditService := NewAuditService()
		_ = auditService.LogLogin(&user.ID, user.Email, false, ipAddress, userAgent, "password verified, MFA required")

		return &models.AuthResponse{
			User:         user,
			AccessToken:  "",
			RefreshToken: "",
			ExpiresIn:    0,
			// Frontend should check for empty tokens and prompt for MFA
		}, nil
	}

	// Generate tokens
	accessToken, refreshToken, expiresIn, err := s.generateTokenPair(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Log successful login (GAP-SEC-008)
	auditService := NewAuditService()
	_ = auditService.LogLogin(&user.ID, user.Email, true, ipAddress, userAgent, "successful login")

	return &models.AuthResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
	}, nil
}

// LoginWithMFA completes login after MFA code verification (GAP-SEC-001)
func (s *AuthService) LoginWithMFA(email, password, mfaCode string, ipAddress, userAgent string, encryption *utils.EncryptionService) (*models.AuthResponse, error) {
	var user models.User
	if err := s.db.Where("email = ? AND is_active = ?", email, true).First(&user).Error; err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Check if MFA is enabled
	if !user.MFAEnabled {
		return nil, fmt.Errorf("MFA is not enabled for this user")
	}

	// Verify MFA code
	mfaSvc := NewMFAService(encryption)
	valid, err := mfaSvc.VerifyMFACode(user.ID.String(), mfaCode)
	if err != nil {
		return nil, err
	}
	if !valid {
		return nil, fmt.Errorf("invalid MFA code")
	}

	// Generate tokens
	accessToken, refreshToken, expiresIn, err := s.generateTokenPair(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Log successful login
	auditService := NewAuditService()
	_ = auditService.LogLogin(&user.ID, user.Email, true, ipAddress, userAgent, "successful login with MFA")

	return &models.AuthResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
	}, nil
}

// LoginWithBackupCode completes login using a backup code (GAP-SEC-001)
func (s *AuthService) LoginWithBackupCode(email, password, backupCode string, ipAddress, userAgent string, encryption *utils.EncryptionService) (*models.AuthResponse, error) {
	var user models.User
	if err := s.db.Where("email = ? AND is_active = ?", email, true).First(&user).Error; err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Check if MFA is enabled
	if !user.MFAEnabled {
		return nil, fmt.Errorf("MFA is not enabled for this user")
	}

	// Verify backup code
	mfaSvc := NewMFAService(encryption)
	valid, err := mfaSvc.VerifyBackupCode(user.ID.String(), backupCode)
	if err != nil {
		return nil, err
	}
	if !valid {
		return nil, fmt.Errorf("invalid backup code")
	}

	// Generate tokens
	accessToken, refreshToken, expiresIn, err := s.generateTokenPair(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Log successful login
	auditService := NewAuditService()
	_ = auditService.LogLogin(&user.ID, user.Email, true, ipAddress, userAgent, "successful login with backup code")

	return &models.AuthResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
	}, nil
}

// RefreshToken generates new tokens from refresh token
func (s *AuthService) RefreshToken(req models.RefreshTokenRequest) (*models.AuthResponse, error) {
	// Verify refresh token
	token, err := jwt.Parse(req.RefreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(config.GetConfig().JWTSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid refresh token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	userIDStr, ok := claims["sub"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid user ID in token")
	}

	tokenType, ok := claims["type"].(string)
	if !ok || tokenType != "refresh" {
		return nil, fmt.Errorf("invalid token type")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format")
	}

	// Check if token is revoked
	var authToken models.AuthToken
	if err := s.db.Where("token = ? AND is_revoked = ?", req.RefreshToken, false).First(&authToken).Error; err != nil {
		return nil, fmt.Errorf("refresh token not found or revoked")
	}

	// Get user
	var user models.User
	if err := s.db.Where("id = ? AND is_active = ?", userID, true).First(&user).Error; err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Revoke old tokens
	if err := s.revokeUserTokens(userID); err != nil {
		return nil, fmt.Errorf("failed to revoke old tokens: %w", err)
	}

	// Generate new tokens
	accessToken, refreshToken, expiresIn, err := s.generateTokenPair(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new tokens: %w", err)
	}

	return &models.AuthResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
	}, nil
}

// ValidateAPIKey validates an API key and returns the associated user
func (s *AuthService) ValidateAPIKey(apiKey string) (*models.User, error) {
	var user models.User
	if err := s.db.Where("api_key = ? AND is_active = ?", apiKey, true).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("invalid API key")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	return &user, nil
}

// ValidateJWT validates a JWT token and returns the associated user
func (s *AuthService) ValidateJWT(tokenString string) (*models.User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(config.GetConfig().JWTSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	userIDStr, ok := claims["sub"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid user ID in token")
	}

	tokenType, ok := claims["type"].(string)
	if !ok || tokenType != "access" {
		return nil, fmt.Errorf("invalid token type")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format")
	}

	var user models.User
	if err := s.db.Where("id = ? AND is_active = ?", userID, true).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	return &user, nil
}

// Logout revokes user tokens
func (s *AuthService) Logout(userID uuid.UUID) error {
	return s.revokeUserTokens(userID)
}

// generateTokenPair generates access and refresh token pair
func (s *AuthService) generateTokenPair(userID uuid.UUID) (string, string, int64, error) {
	conf := config.GetConfig()

	// Set default expiration if not configured
	accessExp := 24  // hours
	refreshExp := 30 // days

	if conf.JWTExpirationHours > 0 {
		accessExp = conf.JWTExpirationHours
	}
	if conf.JWTRefreshDays > 0 {
		refreshExp = conf.JWTRefreshDays
	}

	now := time.Now()
	accessExpiry := now.Add(time.Hour * time.Duration(accessExp))
	refreshExpiry := now.Add(time.Hour * 24 * time.Duration(refreshExp))

	// Generate access token
	accessClaims := jwt.MapClaims{
		"sub":  userID.String(),
		"type": "access",
		"iat":  now.Unix(),
		"exp":  accessExpiry.Unix(),
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(conf.JWTSecret))
	if err != nil {
		return "", "", 0, err
	}

	// Generate refresh token
	refreshClaims := jwt.MapClaims{
		"sub":  userID.String(),
		"type": "refresh",
		"iat":  now.Unix(),
		"exp":  refreshExpiry.Unix(),
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(conf.JWTSecret))
	if err != nil {
		return "", "", 0, err
	}

	// Store tokens in database
	authTokens := []models.AuthToken{
		{
			UserID:    userID,
			TokenType: "access",
			Token:     accessTokenString,
			ExpiresAt: accessExpiry,
		},
		{
			UserID:    userID,
			TokenType: "refresh",
			Token:     refreshTokenString,
			ExpiresAt: refreshExpiry,
		},
	}

	if err := s.db.Create(&authTokens).Error; err != nil {
		return "", "", 0, err
	}

	expiresIn := int64(accessExp * 3600) // Convert hours to seconds
	return accessTokenString, refreshTokenString, expiresIn, nil
}

// revokeUserTokens revokes all tokens for a user
func (s *AuthService) revokeUserTokens(userID uuid.UUID) error {
	return s.db.Model(&models.AuthToken{}).
		Where("user_id = ?", userID).
		Update("is_revoked", true).Error
}

// generateRandomKey generates a random hex key
func generateRandomKey(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// ChangePassword changes a user's password with history tracking (GAP-SEC-002)
func (s *AuthService) ChangePassword(userID uuid.UUID, oldPassword, newPassword, ipAddress, userAgent string) error {
	// Get user
	var user models.User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		// Log failed password change attempt
		auditService := NewAuditService()
		details := &AuditEventDetails{
			Email:  user.Email,
			Reason: "invalid old password",
		}
		_ = auditService.LogEvent(&userID, "password_change_failed", "user", &userID, "failure", ipAddress, userAgent, details)
		return fmt.Errorf("invalid old password")
	}

	// Validate new password strength (GAP-SEC-002)
	if err := utils.ValidatePassword(newPassword, utils.DefaultPasswordPolicy()); err != nil {
		return fmt.Errorf("weak password: %w", err)
	}

	// Check password history - prevent reuse of last 5 passwords (GAP-SEC-002)
	const maxPasswordHistory = 5
	var passwordHistory []models.PasswordHistory
	if err := s.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(maxPasswordHistory).
		Find(&passwordHistory).Error; err != nil {
		return fmt.Errorf("failed to fetch password history: %w", err)
	}

	// Check if new password matches any recent password
	for _, ph := range passwordHistory {
		if err := bcrypt.CompareHashAndPassword([]byte(ph.PasswordHash), []byte(newPassword)); err == nil {
			// Log failed password change attempt
			auditService := NewAuditService()
			details := &AuditEventDetails{
				Email:  user.Email,
				Reason: "password reuse detected (matches password from history)",
			}
			_ = auditService.LogEvent(&userID, "password_change_failed", "user", &userID, "failure", ipAddress, userAgent, details)
			return fmt.Errorf("cannot reuse recent passwords. Please choose a different password")
		}
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update user password
	user.PasswordHash = string(hashedPassword)
	if err := s.db.Save(&user).Error; err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Add new password to history
	newHistory := models.PasswordHistory{
		UserID:       userID,
		PasswordHash: string(hashedPassword),
	}
	if err := s.db.Create(&newHistory).Error; err != nil {
		// Log warning but don't fail the password change
		fmt.Printf("Warning: Failed to create password history for user %s: %v\n", user.Email, err)
	}

	// Clean up old password history (keep only last 5)
	var allHistory []models.PasswordHistory
	if err := s.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&allHistory).Error; err == nil && len(allHistory) > maxPasswordHistory {

		// Delete history entries beyond the limit
		oldHistoryIDs := make([]uuid.UUID, 0)
		for i := maxPasswordHistory; i < len(allHistory); i++ {
			oldHistoryIDs = append(oldHistoryIDs, allHistory[i].ID)
		}
		if len(oldHistoryIDs) > 0 {
			s.db.Where("id IN ?", oldHistoryIDs).Delete(&models.PasswordHistory{})
		}
	}

	// Revoke all existing tokens to force re-login
	if err := s.revokeUserTokens(userID); err != nil {
		fmt.Printf("Warning: Failed to revoke tokens for user %s: %v\n", user.Email, err)
	}

	// Log successful password change
	auditService := NewAuditService()
	_ = auditService.LogPasswordChange(userID, ipAddress, userAgent)

	return nil
}
