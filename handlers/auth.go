package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/dhawalhost/leapmailr/database"
	"github.com/dhawalhost/leapmailr/models"
	"github.com/dhawalhost/leapmailr/service"
	"github.com/gin-gonic/gin"
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

// LoginHandler handles user login
func LoginHandler(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	authService := service.NewAuthService()
	response, err := authService.Login(req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   response,
	})
}

// RefreshTokenHandler handles token refresh
func RefreshTokenHandler(c *gin.Context) {
	var req models.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	authService := service.NewAuthService()
	response, err := authService.RefreshToken(req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   response,
	})
}

// LogoutHandler handles user logout
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
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authService := service.NewAuthService()

		// Check for Authorization header (JWT)
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			// Extract JWT token
			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "Invalid authorization header format",
				})
				c.Abort()
				return
			}

			user, err := authService.ValidateJWT(tokenParts[1])
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "Invalid or expired token",
				})
				c.Abort()
				return
			}

			c.Set("user", user)
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
