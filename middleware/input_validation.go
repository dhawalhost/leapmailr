package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/dhawalhost/leapmailr/utils"

	"github.com/gin-gonic/gin"
)

// ContentTypeValidator middleware ensures requests have valid content-type headers
func ContentTypeValidator() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only validate POST, PUT, PATCH requests
		method := c.Request.Method
		if method != http.MethodPost && method != http.MethodPut && method != http.MethodPatch {
			c.Next()
			return
		}

		contentType := c.GetHeader("Content-Type")
		if contentType == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Content-Type header is required",
			})
			c.Abort()
			return
		}

		// Check if content type is in allowed list
		allowedTypes := append(utils.JSONContentTypes, utils.MultipartContentTypes...)
		if !utils.ValidateContentType(contentType, allowedTypes) {
			c.JSON(http.StatusUnsupportedMediaType, gin.H{
				"error": "Unsupported Content-Type. Use application/json or multipart/form-data",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// InputSanitizer middleware sanitizes all request inputs
func InputSanitizer() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only sanitize JSON requests
		contentType := c.GetHeader("Content-Type")
		if !utils.ValidateContentType(contentType, utils.JSONContentTypes) {
			c.Next()
			return
		}

		// Read the request body
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Failed to read request body",
			})
			c.Abort()
			return
		}

		// Close and restore the body
		c.Request.Body.Close()
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

		// Parse JSON
		var data map[string]interface{}
		if err := json.Unmarshal(body, &data); err != nil {
			// Not valid JSON, let the handler deal with it
			c.Next()
			return
		}

		// Sanitize all string values
		sanitized := sanitizeMap(data)

		// Convert back to JSON
		sanitizedBody, err := json.Marshal(sanitized)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to process request",
			})
			c.Abort()
			return
		}

		// Replace the request body with sanitized version
		c.Request.Body = io.NopCloser(bytes.NewBuffer(sanitizedBody))
		c.Request.ContentLength = int64(len(sanitizedBody))

		c.Next()
	}
}

// sanitizeMap recursively sanitizes all string values in a map
func sanitizeMap(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for key, value := range data {
		switch v := value.(type) {
		case string:
			result[key] = utils.SanitizeInput(v)
		case map[string]interface{}:
			result[key] = sanitizeMap(v)
		case []interface{}:
			result[key] = sanitizeSlice(v)
		default:
			result[key] = value
		}
	}

	return result
}

// sanitizeSlice recursively sanitizes all string values in a slice
func sanitizeSlice(data []interface{}) []interface{} {
	result := make([]interface{}, len(data))

	for i, value := range data {
		switch v := value.(type) {
		case string:
			result[i] = utils.SanitizeInput(v)
		case map[string]interface{}:
			result[i] = sanitizeMap(v)
		case []interface{}:
			result[i] = sanitizeSlice(v)
		default:
			result[i] = value
		}
	}

	return result
}

// XSSProtection middleware adds XSS protection headers
func XSSProtection() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Next()
	}
}

// ValidateEmailAttachments middleware validates email attachment content types and sizes
func ValidateEmailAttachments() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only validate email sending endpoints
		path := c.Request.URL.Path
		if !strings.Contains(path, "/email/send") && !strings.Contains(path, "/email/bulk") {
			c.Next()
			return
		}

		// Read the request body to check attachments
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Failed to read request body",
			})
			c.Abort()
			return
		}

		// Restore the body
		c.Request.Body.Close()
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

		// Parse JSON
		var data map[string]interface{}
		if err := json.Unmarshal(body, &data); err != nil {
			c.Next()
			return
		}

		// Check for attachments
		attachments, ok := data["attachments"].([]interface{})
		if !ok || len(attachments) == 0 {
			c.Next()
			return
		}

		// Validate each attachment
		const maxAttachmentSize = 10 * 1024 * 1024 // 10MB
		totalSize := 0

		for i, att := range attachments {
			attachment, ok := att.(map[string]interface{})
			if !ok {
				continue
			}

			// Check content type
			contentType, _ := attachment["content_type"].(string)
			if contentType != "" && !utils.ValidateContentType(contentType, utils.AllowedAttachmentTypes) {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":      "Invalid attachment content type",
					"attachment": i,
					"type":       contentType,
				})
				c.Abort()
				return
			}

			// Check size
			size, _ := attachment["size"].(float64)
			if size > float64(maxAttachmentSize) {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":      "Attachment exceeds maximum size of 10MB",
					"attachment": i,
					"size":       size,
				})
				c.Abort()
				return
			}

			totalSize += int(size)
		}

		// Check total size
		const maxTotalSize = 25 * 1024 * 1024 // 25MB total
		if totalSize > maxTotalSize {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":      "Total attachment size exceeds maximum of 25MB",
				"total_size": totalSize,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
