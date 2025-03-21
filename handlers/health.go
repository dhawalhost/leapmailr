package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HandleHealthCheck handles the health check endpoint
func HandleHealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
