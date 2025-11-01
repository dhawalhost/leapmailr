package handlers

import (
	"net/http"

	"github.com/dhawalhost/leapmailr/models"
	"github.com/dhawalhost/leapmailr/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CreateCaptchaConfigHandler creates a new CAPTCHA configuration
func CreateCaptchaConfigHandler(c *gin.Context) {
	var req models.CreateCaptchaConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, _ := c.Get("user")
	userData := user.(*models.User)

	captchaService := service.NewCaptchaService()
	response, err := captchaService.CreateCaptchaConfig(req, userData.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// ListCaptchaConfigsHandler lists all CAPTCHA configurations for the user
func ListCaptchaConfigsHandler(c *gin.Context) {
	user, _ := c.Get("user")
	userData := user.(*models.User)

	captchaService := service.NewCaptchaService()
	configs, err := captchaService.ListCaptchaConfigs(userData.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"captcha_configs": configs})
}

// GetCaptchaConfigHandler retrieves a single CAPTCHA configuration
func GetCaptchaConfigHandler(c *gin.Context) {
	configID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid config ID"})
		return
	}

	user, _ := c.Get("user")
	userData := user.(*models.User)

	captchaService := service.NewCaptchaService()
	config, err := captchaService.GetCaptchaConfig(configID, userData.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Configuration not found"})
		return
	}

	c.JSON(http.StatusOK, config)
}

// UpdateCaptchaConfigHandler updates a CAPTCHA configuration
func UpdateCaptchaConfigHandler(c *gin.Context) {
	configID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid config ID"})
		return
	}

	var req models.UpdateCaptchaConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, _ := c.Get("user")
	userData := user.(*models.User)

	captchaService := service.NewCaptchaService()
	response, err := captchaService.UpdateCaptchaConfig(configID, userData.ID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// DeleteCaptchaConfigHandler deletes a CAPTCHA configuration
func DeleteCaptchaConfigHandler(c *gin.Context) {
	configID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid config ID"})
		return
	}

	user, _ := c.Get("user")
	userData := user.(*models.User)

	captchaService := service.NewCaptchaService()
	if err := captchaService.DeleteCaptchaConfig(configID, userData.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "CAPTCHA configuration deleted successfully"})
}
