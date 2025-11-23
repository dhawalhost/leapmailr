package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/dhawalhost/leapmailr/config"
	"github.com/dhawalhost/leapmailr/database"
	"github.com/dhawalhost/leapmailr/models"
	"github.com/dhawalhost/leapmailr/service"
	"github.com/dhawalhost/leapmailr/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Common error messages
const (
	errUserAgent          = "User-Agent"
	errCSRFTokenGenFailed = "Failed to generate CSRF token"
)

// RegisterHandler handles user registration
func RegisterHandler(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	authService := service.NewAuthService()
	response, err := authService.Register(req)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status": "success",
		"data":   response,
	})
}

// LoginHandler handles user login (GAP-SEC-014: Cookie-based auth)
func LoginHandler(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Extract IP address and user agent for audit logging (GAP-SEC-008)
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader(errUserAgent)

	authService := service.NewAuthService()
	response, err := authService.Login(req, ipAddress, userAgent)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Check if MFA is required (empty tokens indicate MFA needed)
	if response.AccessToken == "" {
		c.JSON(http.StatusOK, gin.H{
			"status":      "mfa_required",
			"message":     "MFA verification required",
			"mfa_enabled": true,
		})
		return
	}

	// Generate CSRF token
	csrfService := utils.GetCSRFService()
	userID, _ := uuid.Parse(response.User.ID.String())
	csrfToken, err := csrfService.GenerateToken(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errCSRFTokenGenFailed,
		})
		return
	}

	// Set secure cookies (GAP-SEC-014)
	conf := config.LoadConfig()
	utils.SetAuthCookies(c, response.AccessToken, response.RefreshToken, csrfToken, conf.EnvMode)

	// Return success with user data and tokens (for hybrid cookie + localStorage approach)
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"user":          response.User,
			"access_token":  response.AccessToken,
			"refresh_token": response.RefreshToken,
			"csrf_token":    csrfToken,
		},
	})
}

// LoginWithMFAHandler handles login with MFA code (GAP-SEC-001, GAP-SEC-014)
// POST /api/v1/auth/login-mfa
func LoginWithMFAHandler(c *gin.Context) {
	var req models.MFALoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Extract IP address and user agent for audit logging
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader(errUserAgent)

	// Get encryption service
	encryption, err := utils.NewEncryptionService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to initialize encryption service",
		})
		return
	}

	authService := service.NewAuthService()
	response, err := authService.LoginWithMFA(req.Email, req.Password, req.Code, ipAddress, userAgent, encryption)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Generate CSRF token
	csrfService := utils.GetCSRFService()
	userID, _ := uuid.Parse(response.User.ID.String())
	csrfToken, err := csrfService.GenerateToken(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errCSRFTokenGenFailed,
		})
		return
	}

	// Set secure cookies (GAP-SEC-014)
	conf := config.LoadConfig()
	utils.SetAuthCookies(c, response.AccessToken, response.RefreshToken, csrfToken, conf.EnvMode)

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"user":          response.User,
			"access_token":  response.AccessToken,
			"refresh_token": response.RefreshToken,
			"csrf_token":    csrfToken,
		},
	})
}

// LoginWithBackupCodeHandler handles login with backup code instead of MFA (GAP-SEC-001)
// POST /api/v1/auth/login-backup-code
func LoginWithBackupCodeHandler(c *gin.Context) {
	var req models.MFABackupCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Extract IP address and user agent for audit logging
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader(errUserAgent)

	// Get encryption service
	encryption, err := utils.NewEncryptionService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to initialize encryption service",
		})
		return
	}

	authService := service.NewAuthService()
	response, err := authService.LoginWithBackupCode(req.Email, req.Password, req.BackupCode, ipAddress, userAgent, encryption)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Generate CSRF token
	csrfService := utils.GetCSRFService()
	userID, _ := uuid.Parse(response.User.ID.String())
	csrfToken, err := csrfService.GenerateToken(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errCSRFTokenGenFailed,
		})
		return
	}

	// Set secure cookies (GAP-SEC-014)
	conf := config.LoadConfig()
	utils.SetAuthCookies(c, response.AccessToken, response.RefreshToken, csrfToken, conf.EnvMode)

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"user":          response.User,
			"access_token":  response.AccessToken,
			"refresh_token": response.RefreshToken,
			"csrf_token":    csrfToken,
		},
	})
}

// RefreshTokenHandler handles token refresh (GAP-SEC-014: Cookie-based)
func RefreshTokenHandler(c *gin.Context) {
	// Try to get refresh token from cookie first
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil || refreshToken == "" {
		// Fallback to JSON body for backward compatibility
		var req models.RefreshTokenRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Refresh token required",
			})
			return
		}
		refreshToken = req.RefreshToken
	}

	authService := service.NewAuthService()
	response, err := authService.RefreshToken(models.RefreshTokenRequest{RefreshToken: refreshToken})
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Generate new CSRF token
	csrfService := utils.GetCSRFService()
	userID, _ := uuid.Parse(response.User.ID.String())
	csrfToken, err := csrfService.GenerateToken(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errCSRFTokenGenFailed,
		})
		return
	}

	// Update cookies with new tokens
	conf := config.LoadConfig()
	utils.SetAuthCookies(c, response.AccessToken, response.RefreshToken, csrfToken, conf.EnvMode)

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"user":          response.User,
			"access_token":  response.AccessToken,
			"refresh_token": response.RefreshToken,
			"csrf_token":    csrfToken,
		},
	})
}

// LogoutHandler handles user logout (GAP-SEC-014: Clear cookies and CSRF token)
func LogoutHandler(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	userData := user.(*models.User)
	authService := service.NewAuthService()
	if err := authService.Logout(userData.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to logout",
		})
		return
	}

	// Delete CSRF token from service
	csrfToken, _ := c.Cookie("csrf_token")
	if csrfToken != "" {
		csrfService := utils.GetCSRFService()
		csrfService.DeleteToken(csrfToken)
	}

	// Clear authentication cookies
	utils.DeleteAuthCookies(c)

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Logged out successfully",
	})
}

// ProfileHandler returns the current user's profile
func ProfileHandler(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	userData := user.(*models.User)

	// Remove sensitive information
	userData.PasswordHash = ""
	userData.PrivateKey = ""

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   userData,
	})
}

// AuthMiddleware validates JWT tokens, API keys, or public/private key pairs
// AuthMiddleware validates user authentication via JWT (header or cookie) or API keys (GAP-SEC-014)
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authService := service.NewAuthService()

		// Check for Authorization header first (JWT)
		authHeader := c.GetHeader("Authorization")
		var tokenString string

		if authHeader != "" {
			// Extract JWT token from header
			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "Invalid authorization header format",
				})
				c.Abort()
				return
			}
			tokenString = tokenParts[1]
		} else {
			// Try to get token from cookie (GAP-SEC-014)
			cookieToken, err := c.Cookie("access_token")
			if err == nil && cookieToken != "" {
				tokenString = cookieToken
			}
		}

		// If we have a token (from header or cookie), validate it
		if tokenString != "" {
			user, err := authService.ValidateJWT(tokenString)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "Invalid or expired token",
				})
				c.Abort()
				return
			}

			c.Set("user", user)
			c.Set("userID", user.ID.String())
			c.Set("auth_method", "jwt")
			c.Next()
			return
		}

		// Check for public/private key pair in headers (for SDK usage)
		publicKey := c.GetHeader("X-Public-Key")
		privateKey := c.GetHeader("X-Private-Key")

		// Enhanced authentication: Both public and private keys provided (backend usage)
		if publicKey != "" && privateKey != "" {
			keyPair, err := service.ValidateAPIKeyPair(publicKey, privateKey)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "Invalid or expired API key pair",
				})
				c.Abort()
				return
			}

			// Get user associated with this key pair
			var user models.User
			if err := database.GetDB().Where("id = ?", keyPair.UserID).First(&user).Error; err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "User not found",
				})
				c.Abort()
				return
			}

			c.Set("user", user)
			c.Set("userID", user.ID.String())
			c.Set("auth_method", "api_key_pair_full")
			c.Set("auth_level", "enhanced") // Full API access
			c.Set("api_key_id", keyPair.ID)
			c.Next()
			return
		}

		// Basic authentication: Only public key provided (frontend usage)
		if publicKey != "" {
			keyPair, err := service.ValidatePublicKey(publicKey)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "Invalid or expired public key",
				})
				c.Abort()
				return
			}

			// Get user associated with this key pair
			var user models.User
			if err := database.GetDB().Where("id = ?", keyPair.UserID).First(&user).Error; err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "User not found",
				})
				c.Abort()
				return
			}

			c.Set("user", user)
			c.Set("userID", user.ID.String())
			c.Set("auth_method", "public_key_only")
			c.Set("auth_level", "basic") // Limited to template-based sending
			c.Set("api_key_id", keyPair.ID)

			// Update last used timestamp for public key usage
			now := time.Now()
			database.GetDB().Model(&keyPair).Updates(map[string]interface{}{
				"last_used_at": now,
				"usage_count":  keyPair.UsageCount + 1,
			})

			c.Next()
			return
		}

		// Check for API key in header
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			// Check for API key in query parameter
			apiKey = c.Query("api_key")
		}

		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Missing authentication credentials",
			})
			c.Abort()
			return
		}

		user, err := authService.ValidateAPIKey(apiKey)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid API key",
			})
			c.Abort()
			return
		}

		c.Set("user", user)
		c.Set("userID", user.ID.String())
		c.Set("auth_method", "api_key")
		c.Next()
	}
}

// OptionalAuthMiddleware provides optional authentication
func OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authService := service.NewAuthService()

		// Check for Authorization header (JWT)
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) == 2 && tokenParts[0] == "Bearer" {
				user, err := authService.ValidateJWT(tokenParts[1])
				if err == nil {
					c.Set("user", user)
					c.Set("auth_method", "jwt")
				}
			}
		} else {
			// Check for API key
			apiKey := c.GetHeader("X-API-Key")
			if apiKey == "" {
				apiKey = c.Query("api_key")
			}

			if apiKey != "" {
				user, err := authService.ValidateAPIKey(apiKey)
				if err == nil {
					c.Set("user", user)
					c.Set("auth_method", "api_key")
				}
			}
		}

		c.Next()
	}
}

// GetUserFromContext extracts user from gin context
func GetUserFromContext(c *gin.Context) (*models.User, error) {
	user, exists := c.Get("user")
	if !exists {
		return nil, gin.Error{Err: gin.Error{}, Type: gin.ErrorTypePublic}
	}

	userData, ok := user.(*models.User)
	if !ok {
		return nil, gin.Error{Err: gin.Error{}, Type: gin.ErrorTypePublic}
	}

	return userData, nil
}

// RequireUser middleware that requires authenticated user
func RequireUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, err := GetUserFromContext(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// ChangePasswordHandler handles password change requests (GAP-SEC-002)
func ChangePasswordHandler(c *gin.Context) {
	// Get authenticated user
	user, err := GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}

	// Parse request
	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Get IP address and user agent
	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	// Change password
	authService := service.NewAuthService()
	if err := authService.ChangePassword(user.ID, req.OldPassword, req.NewPassword, ipAddress, userAgent); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Password changed successfully. Please login again with your new password.",
	})
}
