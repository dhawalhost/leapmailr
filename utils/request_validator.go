package utils

import (
	"net"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

// CustomValidator holds custom validation functions
type CustomValidator struct {
	Validator *validator.Validate
}

// NewCustomValidator creates a new validator with custom rules
func NewCustomValidator() *CustomValidator {
	v := validator.New()

	// Register custom validators
	_ = v.RegisterValidation("mxlookup", validateEmailMX)
	_ = v.RegisterValidation("noxss", validateNoXSS)
	_ = v.RegisterValidation("nosqli", validateNoSQLi)
	_ = v.RegisterValidation("safepath", validateSafePath)
	_ = v.RegisterValidation("alphanumunicode", validateAlphaNumUnicode)

	return &CustomValidator{Validator: v}
}

// validateEmailMX validates that an email domain has valid MX records
func validateEmailMX(fl validator.FieldLevel) bool {
	email := fl.Field().String()
	if email == "" {
		return true // Let required tag handle empty strings
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}

	domain := parts[1]
	mx, err := net.LookupMX(domain)
	if err != nil || len(mx) == 0 {
		return false
	}

	return true
}

// validateNoXSS checks for common XSS patterns
func validateNoXSS(fl validator.FieldLevel) bool {
	str := fl.Field().String()
	if str == "" {
		return true
	}

	// Common XSS patterns
	xssPatterns := []string{
		`<script`,
		`javascript:`,
		`onerror=`,
		`onclick=`,
		`onload=`,
		`<iframe`,
		`<embed`,
		`<object`,
		`eval\(`,
		`expression\(`,
	}

	lowerStr := strings.ToLower(str)
	for _, pattern := range xssPatterns {
		if strings.Contains(lowerStr, pattern) {
			return false
		}
	}

	// Check for encoded script tags
	encodedPatterns := []string{
		`%3Cscript`,
		`&lt;script`,
		`&#60;script`,
	}

	for _, pattern := range encodedPatterns {
		if strings.Contains(lowerStr, pattern) {
			return false
		}
	}

	return true
}

// validateNoSQLi checks for common SQL injection patterns
func validateNoSQLi(fl validator.FieldLevel) bool {
	str := fl.Field().String()
	if str == "" {
		return true
	}

	// Common SQL injection patterns
	sqlPatterns := []string{
		`'`,
		`"`,
		`--`,
		`;`,
		`/*`,
		`*/`,
		`xp_`,
		`sp_`,
		`exec`,
		`execute`,
		`union.*select`,
		`insert.*into`,
		`delete.*from`,
		`drop.*table`,
		`update.*set`,
	}

	lowerStr := strings.ToLower(str)
	for _, pattern := range sqlPatterns {
		matched, _ := regexp.MatchString(pattern, lowerStr)
		if matched {
			return false
		}
	}

	return true
}

// validateSafePath ensures no path traversal attempts
func validateSafePath(fl validator.FieldLevel) bool {
	str := fl.Field().String()
	if str == "" {
		return true
	}

	// Path traversal patterns
	dangerousPatterns := []string{
		"..",
		"./",
		"~/",
		"\\",
		"%2e%2e",
		"%252e",
		"..%2f",
		"..%5c",
	}

	lowerStr := strings.ToLower(str)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(lowerStr, pattern) {
			return false
		}
	}

	return true
}

// validateAlphaNumUnicode allows alphanumeric characters, spaces, and common punctuation
func validateAlphaNumUnicode(fl validator.FieldLevel) bool {
	str := fl.Field().String()
	if str == "" {
		return true
	}

	// Allow letters (any language), numbers, spaces, and common punctuation
	pattern := `^[\p{L}\p{N}\s\-_.,!?@#&()'"]+$`
	matched, err := regexp.MatchString(pattern, str)
	if err != nil {
		return false
	}

	return matched
}

// ValidateContentType checks if content type is in allowed list
func ValidateContentType(contentType string, allowedTypes []string) bool {
	contentType = strings.ToLower(strings.TrimSpace(contentType))

	// Extract the main type (before semicolon)
	if idx := strings.Index(contentType, ";"); idx != -1 {
		contentType = strings.TrimSpace(contentType[:idx])
	}

	for _, allowed := range allowedTypes {
		if contentType == strings.ToLower(allowed) {
			return true
		}
	}

	return false
}

// Common content type lists
var (
	JSONContentTypes = []string{
		"application/json",
		"application/json; charset=utf-8",
	}

	MultipartContentTypes = []string{
		"multipart/form-data",
	}

	ImageContentTypes = []string{
		"image/jpeg",
		"image/jpg",
		"image/png",
		"image/gif",
		"image/webp",
		"image/svg+xml",
	}

	DocumentContentTypes = []string{
		"application/pdf",
		"application/msword",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/vnd.ms-excel",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		"text/plain",
		"text/csv",
	}

	AllowedAttachmentTypes = append(append([]string{}, ImageContentTypes...), DocumentContentTypes...)
)
