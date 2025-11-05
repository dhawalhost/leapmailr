package handlers

import (
	"net/http"

	"github.com/dhawalhost/leapmailr/models"
	"github.com/dhawalhost/leapmailr/service"
	"github.com/dhawalhost/leapmailr/utils"
	"github.com/gin-gonic/gin"
)

var mfaService *service.MFAService

// InitMFAService initializes the MFA service
func InitMFAService(encryption *utils.EncryptionService) {
	mfaService = service.NewMFAService(encryption)
}

// SetupMFAHandler initiates MFA setup for the authenticated user
// POST /api/v1/mfa/setup
func SetupMFAHandler(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req models.MFASetupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := mfaService.SetupMFA(userID, req.Password)
	if err != nil {
		if err == service.ErrMFAAlreadyEnabled {
			c.JSON(http.StatusConflict, gin.H{"error": "MFA is already enabled"})
			return
		}
		if err == service.ErrInvalidPassword {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid password"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to setup MFA"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// VerifyMFASetupHandler verifies the TOTP code and enables MFA
// POST /api/v1/mfa/verify-setup
func VerifyMFASetupHandler(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req models.MFAVerifySetupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := mfaService.VerifyMFASetup(userID, req.Code); err != nil {
		if err == service.ErrMFAAlreadyEnabled {
			c.JSON(http.StatusConflict, gin.H{"error": "MFA is already enabled"})
			return
		}
		if err == service.ErrInvalidMFACode {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid MFA code"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify MFA setup"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "MFA enabled successfully"})
}

// DisableMFAHandler disables MFA for the authenticated user
// POST /api/v1/mfa/disable
func DisableMFAHandler(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req models.MFADisableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := mfaService.DisableMFA(userID, req.Password, req.Code); err != nil {
		if err == service.ErrMFANotEnabled {
			c.JSON(http.StatusBadRequest, gin.H{"error": "MFA is not enabled"})
			return
		}
		if err == service.ErrInvalidPassword {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid password"})
			return
		}
		if err == service.ErrInvalidMFACode {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid MFA code"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to disable MFA"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "MFA disabled successfully"})
}

// RegenerateBackupCodesHandler generates new backup codes
// POST /api/v1/mfa/regenerate-backup-codes
func RegenerateBackupCodesHandler(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req models.MFARegenerateBackupCodesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	backupCodes, err := mfaService.RegenerateBackupCodes(userID, req.Password, req.Code)
	if err != nil {
		if err == service.ErrMFANotEnabled {
			c.JSON(http.StatusBadRequest, gin.H{"error": "MFA is not enabled"})
			return
		}
		if err == service.ErrInvalidPassword {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid password"})
			return
		}
		if err == service.ErrInvalidMFACode {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid MFA code"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to regenerate backup codes"})
		return
	}

	c.JSON(http.StatusOK, models.MFARegenerateBackupCodesResponse{
		BackupCodes: backupCodes,
	})
}

// GetMFAStatusHandler returns the MFA status for the authenticated user
// GET /api/v1/mfa/status
func GetMFAStatusHandler(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	enabled, backupCodesCount, err := mfaService.GetMFAStatus(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get MFA status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"enabled":            enabled,
		"backup_codes_count": backupCodesCount,
	})
}
