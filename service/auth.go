package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/dhawalhost/leapmailr/config"
	"github.com/dhawalhost/leapmailr/database"
	"github.com/dhawalhost/leapmailr/models"
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
	// Check if user already exists
	var existingUser models.User
	if err := s.db.Debug().Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
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

	// Create user
	user := models.User{
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		APIKey:       apiKey,
		PrivateKey:   privateKey,
		PlanType:     "free",
		IsActive:     true,
	}

	if err := s.db.Create(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
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

// Login authenticates a user
func (s *AuthService) Login(req models.LoginRequest) (*models.AuthResponse, error) {
	var user models.User
	if err := s.db.Where("email = ? AND is_active = ?", req.Email, true).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("invalid credentials")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
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
