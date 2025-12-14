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
		_ = c.Request.Body.Close()
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body)) // Parse JSON
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
		if !isEmailEndpoint(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Parse request body
		body, data, err := parseRequestBody(c)
		if err != nil {
			abortWithError(c, http.StatusBadRequest, "Failed to read request body")
			return
		}

		// Restore body for downstream handlers
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

		// Extract and validate attachments
		if err := validateAttachments(c, data); err != nil {
			return // Error already sent
		}

		c.Next()
	}
}

// isEmailEndpoint checks if the path is an email sending endpoint
func isEmailEndpoint(path string) bool {
	return strings.Contains(path, "/email/send") || strings.Contains(path, "/email/bulk")
}

// parseRequestBody reads and parses the JSON request body
func parseRequestBody(c *gin.Context) ([]byte, map[string]interface{}, error) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil, nil, err
	}

	_ = c.Request.Body.Close()

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		// Not valid JSON - let handler deal with it
		return body, nil, nil
	}

	return body, data, nil
}

// validateAttachments validates attachment sizes and content types
func validateAttachments(c *gin.Context, data map[string]interface{}) error {
	if data == nil {
		return nil
	}

	// Extract attachments
	attachments, ok := data["attachments"].([]interface{})
	if !ok || len(attachments) == 0 {
		return nil
	}

	// Validate each attachment
	totalSize := 0
	for i, att := range attachments {
		attachment, ok := att.(map[string]interface{})
		if !ok {
			continue
		}

		// Validate content type
		if err := validateAttachmentContentType(c, attachment, i); err != nil {
			return err
		}

		// Validate size
		size, sizeErr := validateAttachmentSize(c, attachment, i)
		if sizeErr != nil {
			return sizeErr
		}

		totalSize += size
	}

	// Validate total size
	return validateTotalSize(c, totalSize)
}

// validateAttachmentContentType validates a single attachment's content type
func validateAttachmentContentType(c *gin.Context, attachment map[string]interface{}, index int) error {
	contentType, _ := attachment["content_type"].(string)
	if contentType != "" && !utils.ValidateContentType(contentType, utils.AllowedAttachmentTypes) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Invalid attachment content type",
			"attachment": index,
			"type":       contentType,
		})
		c.Abort()
		return &validationError{message: "invalid content type"}
	}
	return nil
}

// validateAttachmentSize validates a single attachment's size
func validateAttachmentSize(c *gin.Context, attachment map[string]interface{}, index int) (int, error) {
	const maxAttachmentSize = 10 * 1024 * 1024 // 10MB

	size, _ := attachment["size"].(float64)
	if size > float64(maxAttachmentSize) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Attachment exceeds maximum size of 10MB",
			"attachment": index,
			"size":       size,
		})
		c.Abort()
		return 0, &validationError{message: "size too large"}
	}

	return int(size), nil
}

// validateTotalSize validates the total size of all attachments
func validateTotalSize(c *gin.Context, totalSize int) error {
	const maxTotalSize = 25 * 1024 * 1024 // 25MB total

	if totalSize > maxTotalSize {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Total attachment size exceeds maximum of 25MB",
			"total_size": totalSize,
		})
		c.Abort()
		return &validationError{message: "total size too large"}
	}

	return nil
}

// validationError represents a validation error
type validationError struct {
	message string
}

func (e *validationError) Error() string {
	return e.message
}

// abortWithError aborts the request with an error
func abortWithError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, gin.H{
		"error": message,
	})
	c.Abort()
}
