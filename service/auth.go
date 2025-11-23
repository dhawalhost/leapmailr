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

// Auth error messages and query constants
const (
	errFailedGenTokens    = "failed to generate tokens: %w"
	errUserNotFound       = "user not found"
	errInvalidCredentials = "invalid credentials"
	queryEmailIsActive    = "email = ? AND is_active = ?"
	queryUserID           = "user_id = ?"
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
		return nil, fmt.Errorf(errFailedGenTokens, err)
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
	// Find and validate user
	user, err := s.findActiveUser(req.Email, ipAddress, userAgent)
	if err != nil {
		return nil, err
	}

	// Check account lockout
	if err := s.checkAccountLockout(&user, ipAddress, userAgent); err != nil {
		return nil, err
	}

	// Update last login attempt
	now := time.Now()
	user.LastLoginAttempt = &now

	// Verify password
	if err := s.verifyPassword(&user, req.Password, now, ipAddress, userAgent); err != nil {
		return nil, err
	}

	// Handle successful login
	return s.handleSuccessfulLogin(&user, now, ipAddress, userAgent)
}

// findActiveUser finds an active user by email
func (s *AuthService) findActiveUser(email, ipAddress, userAgent string) (models.User, error) {
	var user models.User
	if err := s.db.Where(queryEmailIsActive, email, true).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			auditService := NewAuditService()
			_ = auditService.LogLogin(nil, email, false, ipAddress, userAgent, errUserNotFound)
			return user, fmt.Errorf(errInvalidCredentials)
		}
		return user, fmt.Errorf("database error: %w", err)
	}
	return user, nil
}

// checkAccountLockout checks if account is locked
func (s *AuthService) checkAccountLockout(user *models.User, ipAddress, userAgent string) error {
	now := time.Now()
	if user.LockedUntil != nil && user.LockedUntil.After(now) {
		remainingMinutes := int(time.Until(*user.LockedUntil).Minutes())
		auditService := NewAuditService()
		_ = auditService.LogLogin(&user.ID, user.Email, false, ipAddress, userAgent,
			fmt.Sprintf("account locked for %d more minutes", remainingMinutes))
		return fmt.Errorf("account is locked due to too many failed login attempts. Try again in %d minutes", remainingMinutes)
	}
	return nil
}

// verifyPassword verifies the user password and handles failed attempts
func (s *AuthService) verifyPassword(user *models.User, password string, now time.Time, ipAddress, userAgent string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return s.handleFailedLogin(user, now, ipAddress, userAgent)
	}
	return nil
}

// handleFailedLogin handles failed login attempts and account locking
func (s *AuthService) handleFailedLogin(user *models.User, now time.Time, ipAddress, userAgent string) error {
	const maxFailedAttempts = 5
	const lockoutDurationMinutes = 15

	user.FailedLoginAttempts++

	// Lock account after max failed attempts
	if user.FailedLoginAttempts >= maxFailedAttempts {
		return s.lockAccount(user, now, lockoutDurationMinutes, ipAddress, userAgent)
	}

	// Save failed attempt count
	if err := s.db.Save(user).Error; err != nil {
		fmt.Printf("Warning: Failed to update failed login count for user %s: %v\n", user.Email, err)
	}

	// Log failed login
	auditService := NewAuditService()
	_ = auditService.LogLogin(&user.ID, user.Email, false, ipAddress, userAgent,
		fmt.Sprintf("invalid password (attempt %d/%d)", user.FailedLoginAttempts, maxFailedAttempts))

	return fmt.Errorf(errInvalidCredentials)
}

// lockAccount locks the user account after too many failed attempts
func (s *AuthService) lockAccount(user *models.User, now time.Time, lockoutMinutes int, ipAddress, userAgent string) error {
	lockoutUntil := now.Add(time.Duration(lockoutMinutes) * time.Minute)
	user.LockedUntil = &lockoutUntil

	if err := s.db.Save(user).Error; err != nil {
		fmt.Printf("Warning: Failed to lock account for user %s: %v\n", user.Email, err)
	}

	auditService := NewAuditService()
	_ = auditService.LogAccountLocked(user.ID, user.Email,
		fmt.Sprintf("locked for %d minutes after %d failed attempts", lockoutMinutes, user.FailedLoginAttempts),
		ipAddress, userAgent)

	return fmt.Errorf("account locked due to too many failed login attempts. Try again in %d minutes", lockoutMinutes)
}

// handleSuccessfulLogin handles successful login and token generation
func (s *AuthService) handleSuccessfulLogin(user *models.User, now time.Time, ipAddress, userAgent string) (*models.AuthResponse, error) {
	// Reset failed attempts
	user.FailedLoginAttempts = 0
	user.LockedUntil = nil
	user.LastLoginSuccess = &now

	if err := s.db.Save(user).Error; err != nil {
		fmt.Printf("Warning: Failed to reset failed login count for user %s: %v\n", user.Email, err)
	}

	// Check if MFA is enabled
	if user.MFAEnabled {
		return s.handleMFARequired(user, ipAddress, userAgent)
	}

	// Generate tokens
	return s.generateAuthResponse(user, ipAddress, userAgent, "successful login")
}

// handleMFARequired handles login when MFA is required
func (s *AuthService) handleMFARequired(user *models.User, ipAddress, userAgent string) (*models.AuthResponse, error) {
	auditService := NewAuditService()
	_ = auditService.LogLogin(&user.ID, user.Email, false, ipAddress, userAgent, "password verified, MFA required")

	return &models.AuthResponse{
		User:         *user,
		AccessToken:  "",
		RefreshToken: "",
		ExpiresIn:    0,
	}, nil
}

// generateAuthResponse generates tokens and creates auth response
func (s *AuthService) generateAuthResponse(user *models.User, ipAddress, userAgent, auditMessage string) (*models.AuthResponse, error) {
	accessToken, refreshToken, expiresIn, err := s.generateTokenPair(user.ID)
	if err != nil {
		return nil, fmt.Errorf(errFailedGenTokens, err)
	}

	auditService := NewAuditService()
	_ = auditService.LogLogin(&user.ID, user.Email, true, ipAddress, userAgent, auditMessage)

	return &models.AuthResponse{
		User:         *user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
	}, nil
}

// LoginWithMFA completes login after MFA code verification (GAP-SEC-001)
func (s *AuthService) LoginWithMFA(email, password, mfaCode string, ipAddress, userAgent string, encryption *utils.EncryptionService) (*models.AuthResponse, error) {
	var user models.User
	if err := s.db.Where(queryEmailIsActive, email, true).First(&user).Error; err != nil {
		return nil, fmt.Errorf(errInvalidCredentials)
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, fmt.Errorf(errInvalidCredentials)
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
		return nil, fmt.Errorf(errFailedGenTokens, err)
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
	if err := s.db.Where(queryEmailIsActive, email, true).First(&user).Error; err != nil {
		return nil, fmt.Errorf(errInvalidCredentials)
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, fmt.Errorf(errInvalidCredentials)
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
		return nil, fmt.Errorf(errFailedGenTokens, err)
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
	// Parse and validate refresh token
	claims, err := s.parseRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, err
	}

	// Extract user ID
	userID, err := s.extractUserIDFromClaims(claims)
	if err != nil {
		return nil, err
	}

	// Verify token not revoked
	if err := s.verifyTokenNotRevoked(req.RefreshToken); err != nil {
		return nil, err
	}

	// Get active user
	user, err := s.getActiveUser(userID)
	if err != nil {
		return nil, err
	}

	// Revoke old tokens and generate new ones
	if err := s.revokeUserTokens(userID); err != nil {
		return nil, fmt.Errorf("failed to revoke old tokens: %w", err)
	}

	accessToken, refreshToken, expiresIn, err := s.generateTokenPair(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new tokens: %w", err)
	}

	return &models.AuthResponse{
		User:         *user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
	}, nil
}

// parseRefreshToken parses and validates the JWT refresh token
func (s *AuthService) parseRefreshToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
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

	return claims, nil
}

// extractUserIDFromClaims extracts and validates user ID from JWT claims
func (s *AuthService) extractUserIDFromClaims(claims jwt.MapClaims) (uuid.UUID, error) {
	userIDStr, ok := claims["sub"].(string)
	if !ok {
		return uuid.Nil, fmt.Errorf("invalid user ID in token")
	}

	tokenType, ok := claims["type"].(string)
	if !ok || tokenType != "refresh" {
		return uuid.Nil, fmt.Errorf("invalid token type")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid user ID format")
	}

	return userID, nil
}

// verifyTokenNotRevoked checks if the token has been revoked
func (s *AuthService) verifyTokenNotRevoked(token string) error {
	var authToken models.AuthToken
	if err := s.db.Where("token = ? AND is_revoked = ?", token, false).First(&authToken).Error; err != nil {
		return fmt.Errorf("refresh token not found or revoked")
	}
	return nil
}

// getActiveUser retrieves an active user by ID
func (s *AuthService) getActiveUser(userID uuid.UUID) (*models.User, error) {
	var user models.User
	if err := s.db.Where("id = ? AND is_active = ?", userID, true).First(&user).Error; err != nil {
		return nil, fmt.Errorf(errUserNotFound)
	}
	return &user, nil
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
			return nil, fmt.Errorf(errUserNotFound)
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
		Where(queryUserID, userID).
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
	if err := s.verifyOldPassword(&user, oldPassword, ipAddress, userAgent); err != nil {
		return err
	}

	// Validate new password strength (GAP-SEC-002)
	if err := utils.ValidatePassword(newPassword, utils.DefaultPasswordPolicy()); err != nil {
		return fmt.Errorf("weak password: %w", err)
	}

	// Check password history
	if err := s.checkPasswordHistory(userID, &user, newPassword, ipAddress, userAgent); err != nil {
		return err
	}

	// Hash and update password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user.PasswordHash = string(hashedPassword)
	if err := s.db.Save(&user).Error; err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Update password history
	s.updatePasswordHistory(userID, &user, hashedPassword)

	// Revoke all existing tokens to force re-login
	if err := s.revokeUserTokens(userID); err != nil {
		fmt.Printf("Warning: Failed to revoke tokens for user %s: %v\n", user.Email, err)
	}

	// Log successful password change
	auditService := NewAuditService()
	_ = auditService.LogPasswordChange(userID, ipAddress, userAgent)

	return nil
}

func (s *AuthService) verifyOldPassword(user *models.User, oldPassword, ipAddress, userAgent string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		auditService := NewAuditService()
		details := &AuditEventDetails{
			Email:  user.Email,
			Reason: "invalid old password",
		}
		_ = auditService.LogEvent(&AuditEventParams{
			UserID:     &user.ID,
			Action:     "password_change_failed",
			Resource:   "user",
			ResourceID: &user.ID,
			Status:     "failure",
			IPAddress:  ipAddress,
			UserAgent:  userAgent,
			Details:    details,
		})
		return fmt.Errorf("invalid old password")
	}
	return nil
}

func (s *AuthService) checkPasswordHistory(userID uuid.UUID, user *models.User, newPassword, ipAddress, userAgent string) error {
	const maxPasswordHistory = 5
	var passwordHistory []models.PasswordHistory
	if err := s.db.Where(queryUserID, userID).
		Order("created_at DESC").
		Limit(maxPasswordHistory).
		Find(&passwordHistory).Error; err != nil {
		return fmt.Errorf("failed to fetch password history: %w", err)
	}

	// Check if new password matches any recent password
	for _, ph := range passwordHistory {
		if err := bcrypt.CompareHashAndPassword([]byte(ph.PasswordHash), []byte(newPassword)); err == nil {
			auditService := NewAuditService()
			details := &AuditEventDetails{
				Email:  user.Email,
				Reason: "password reuse detected (matches password from history)",
			}
			_ = auditService.LogEvent(&AuditEventParams{
				UserID:     &userID,
				Action:     "password_change_failed",
				Resource:   "user",
				ResourceID: &userID,
				Status:     "failure",
				IPAddress:  ipAddress,
				UserAgent:  userAgent,
				Details:    details,
			})
			return fmt.Errorf("cannot reuse recent passwords. Please choose a different password")
		}
	}
	return nil
}

func (s *AuthService) updatePasswordHistory(userID uuid.UUID, user *models.User, hashedPassword []byte) {
	const maxPasswordHistory = 5

	// Add new password to history
	newHistory := models.PasswordHistory{
		UserID:       userID,
		PasswordHash: string(hashedPassword),
	}
	if err := s.db.Create(&newHistory).Error; err != nil {
		fmt.Printf("Warning: Failed to create password history for user %s: %v\n", user.Email, err)
	}

	// Clean up old password history (keep only last 5)
	var allHistory []models.PasswordHistory
	if err := s.db.Where(queryUserID, userID).
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
}
