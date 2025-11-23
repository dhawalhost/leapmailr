package utils

import (
	"errors"
	"fmt"
	"net/mail"
	"net/url"
	"os"
	"regexp"
	"strings"
	"unicode"
)

// Common weak passwords to reject
var commonPasswords = map[string]bool{
	"password":    true,
	"password123": true,
	"123456":      true,
	"12345678":    true,
	"qwerty":      true,
	"abc123":      true,
	"password1":   true,
	"admin":       true,
	"admin123":    true,
	"letmein":     true,
	"welcome":     true,
	"monkey":      true,
	"dragon":      true,
	"master":      true,
	"sunshine":    true,
	"princess":    true,
	"football":    true,
	"iloveyou":    true,
	"superman":    true,
	"trustno1":    true,
	"baseball":    true,
	"batman":      true,
	"access":      true,
	"shadow":      true,
	"michael":     true,
	"jennifer":    true,
	"111111":      true,
	"000000":      true,
	"654321":      true,
	"passw0rd":    true,
	"adminadmin":  true,
	"rootroot":    true,
}

// PasswordPolicy defines password requirements
type PasswordPolicy struct {
	MinLength      int
	RequireUpper   bool
	RequireLower   bool
	RequireNumber  bool
	RequireSpecial bool
}

// DefaultPasswordPolicy returns SOC 2 compliant password policy
func DefaultPasswordPolicy() PasswordPolicy {
	return PasswordPolicy{
		MinLength:      12,
		RequireUpper:   true,
		RequireLower:   true,
		RequireNumber:  true,
		RequireSpecial: true,
	}
}

// ValidatePassword validates a password against the policy
func ValidatePassword(password string, policy PasswordPolicy) error {
	if len(password) < policy.MinLength {
		return errors.New("password must be at least 12 characters long")
	}

	if len(password) > 128 {
		return errors.New("password must not exceed 128 characters")
	}

	// Check for common weak passwords
	lowerPassword := strings.ToLower(password)
	if commonPasswords[lowerPassword] {
		return errors.New("password is too common and easily guessable")
	}

	// Check if password contains variations of common words
	for commonPwd := range commonPasswords {
		if strings.Contains(lowerPassword, commonPwd) && len(commonPwd) > 5 {
			return errors.New("password contains common words that are easily guessable")
		}
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if policy.RequireUpper && !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}

	if policy.RequireLower && !hasLower {
		return errors.New("password must contain at least one lowercase letter")
	}

	if policy.RequireNumber && !hasNumber {
		return errors.New("password must contain at least one number")
	}

	if policy.RequireSpecial && !hasSpecial {
		return errors.New("password must contain at least one special character (!@#$%^&*()_+-=[]{}|;:,.<>?)")
	}

	// Check for sequential characters (123, abc, etc.)
	if hasSequentialChars(password) {
		return errors.New("password must not contain sequential characters (e.g., 123, abc)")
	}

	// Check for repeated characters (aaa, 111, etc.)
	if hasRepeatedChars(password, 3) {
		return errors.New("password must not contain more than 2 repeated characters")
	}

	return nil
}

// hasSequentialChars checks if password contains sequential characters
func hasSequentialChars(password string) bool {
	sequential := []string{
		"012", "123", "234", "345", "456", "567", "678", "789",
		"abc", "bcd", "cde", "def", "efg", "fgh", "ghi", "hij",
		"ijk", "jkl", "klm", "lmn", "mno", "nop", "opq", "pqr",
		"qrs", "rst", "stu", "tuv", "uvw", "vwx", "wxy", "xyz",
	}

	lowerPassword := strings.ToLower(password)
	for _, seq := range sequential {
		if strings.Contains(lowerPassword, seq) {
			return true
		}
	}
	return false
}

// hasRepeatedChars checks if password contains repeated characters
func hasRepeatedChars(password string, maxRepeat int) bool {
	if len(password) < maxRepeat {
		return false
	}

	for i := 0; i < len(password)-maxRepeat+1; i++ {
		char := password[i]
		repeated := true
		for j := 1; j < maxRepeat; j++ {
			if password[i+j] != char {
				repeated = false
				break
			}
		}
		if repeated {
			return true
		}
	}
	return false
}

func IsValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

// ValidateEmail validates email format with detailed error
func ValidateEmail(email string) error {
	if email == "" {
		return errors.New("email is required")
	}

	if len(email) > 255 {
		return errors.New("email must not exceed 255 characters")
	}

	if !IsValidEmail(email) {
		return errors.New("invalid email format")
	}

	return nil
}

// SanitizeInput sanitizes user input to prevent XSS
func SanitizeInput(input string) string {
	// Trim whitespace
	input = strings.TrimSpace(input)

	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")

	// Limit length to prevent buffer overflow
	if len(input) > 10000 {
		input = input[:10000]
	}

	return input
}

// ValidateName validates first/last name
func ValidateName(name string) error {
	if name == "" {
		return errors.New("name is required")
	}

	if len(name) > 100 {
		return errors.New("name must not exceed 100 characters")
	}

	// Allow letters, spaces, hyphens, apostrophes
	nameRegex := regexp.MustCompile(`^[a-zA-Z\s'-]+$`)
	if !nameRegex.MatchString(name) {
		return errors.New("name contains invalid characters")
	}

	return nil
}

// GetBaseURL returns the base URL for the application from environment
func GetBaseURL() string {
	// Try to get from environment variable
	baseURL := os.Getenv("BASE_URL")
	if baseURL != "" {
		return strings.TrimSuffix(baseURL, "/")
	}

	// Default to localhost for development
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return "http://localhost:" + port
}

// ValidateRedirectURL validates that a URL is safe to redirect to
// Prevents open redirect vulnerabilities
func ValidateRedirectURL(urlStr string) error {
	if urlStr == "" {
		return errors.New("URL is required")
	}

	// Check URL length
	if len(urlStr) > 2048 {
		return errors.New("URL too long")
	}

	// Parse URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return errors.New("invalid URL format")
	}

	// Must have a scheme (http or https)
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return errors.New("URL must use http or https protocol")
	}

	// Must have a host
	if parsedURL.Host == "" {
		return errors.New("URL must have a valid host")
	}

	// Prevent localhost/internal redirects (security measure)
	host := strings.ToLower(parsedURL.Hostname())
	if host == "localhost" || host == "127.0.0.1" || host == "0.0.0.0" ||
		strings.HasPrefix(host, "192.168.") ||
		strings.HasPrefix(host, "10.") ||
		strings.HasPrefix(host, "172.16.") ||
		host == "::1" {
		return errors.New("redirects to internal/local addresses are not allowed")
	}

	// Prevent javascript: data: and other dangerous schemes
	if strings.Contains(strings.ToLower(urlStr), "javascript:") ||
		strings.Contains(strings.ToLower(urlStr), "data:") ||
		strings.Contains(strings.ToLower(urlStr), "vbscript:") {
		return errors.New("URL contains dangerous content")
	}

	return nil
}

// IsAllowedDomain checks if a domain is in the allowed list
// This should be configured based on your application's needs
func IsAllowedDomain(urlStr string, allowedDomains []string) (bool, error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return false, err
	}

	host := strings.ToLower(parsedURL.Hostname())

	// If no allowed domains specified, use validation only
	if len(allowedDomains) == 0 {
		return true, nil
	}

	// Check if domain matches any allowed domain
	for _, allowed := range allowedDomains {
		allowed = strings.ToLower(strings.TrimSpace(allowed))
		if host == allowed || strings.HasSuffix(host, "."+allowed) {
			return true, nil
		}
	}

	return false, nil
}

// SanitizeHostHeader validates and sanitizes the Host header
// Prevents host header injection attacks
func SanitizeHostHeader(host string) (string, error) {
	if host == "" {
		return "", errors.New("empty host header")
	}

	// Remove port if present for validation
	hostname := host
	if idx := strings.LastIndex(host, ":"); idx != -1 {
		hostname = host[:idx]
	}

	// Validate hostname format
	// Must be valid domain or IP address
	if hostname == "" {
		return "", errors.New("invalid host header")
	}

	// Block localhost and internal addresses
	lower := strings.ToLower(hostname)
	if lower == "localhost" || lower == "127.0.0.1" || lower == "0.0.0.0" ||
		strings.HasPrefix(lower, "192.168.") ||
		strings.HasPrefix(lower, "10.") ||
		strings.HasPrefix(lower, "172.16.") ||
		lower == "::1" {
		return "", errors.New("localhost and internal addresses not allowed")
	}

	// Basic validation: only allow alphanumeric, dots, hyphens, and colons (for port)
	validHost := regexp.MustCompile(`^[a-zA-Z0-9\.\-:]+$`)
	if !validHost.MatchString(host) {
		return "", errors.New("invalid characters in host header")
	}

	return host, nil
}

// SanitizeRequestURI validates and sanitizes the request URI
// Prevents URI injection attacks
func SanitizeRequestURI(uri string) (string, error) {
	if uri == "" {
		uri = "/"
	}

	// URI must start with /
	if !strings.HasPrefix(uri, "/") {
		return "", errors.New("invalid URI format")
	}

	// Prevent absolute URLs in URI (which would be an open redirect)
	if strings.Contains(uri, "://") {
		return "", errors.New("absolute URLs not allowed in URI")
	}

	// Check for suspicious patterns
	lower := strings.ToLower(uri)
	if strings.Contains(lower, "javascript:") ||
		strings.Contains(lower, "data:") ||
		strings.Contains(lower, "vbscript:") {
		return "", errors.New("dangerous content in URI")
	}

	// Limit length
	if len(uri) > 2048 {
		return "", errors.New("URI too long")
	}

	return uri, nil
}

// BuildSecureHTTPSURL safely constructs an HTTPS URL from validated components
func BuildSecureHTTPSURL(host, uri string) (string, error) {
	// Validate and sanitize host
	cleanHost, err := SanitizeHostHeader(host)
	if err != nil {
		return "", fmt.Errorf("invalid host: %w", err)
	}

	// Validate and sanitize URI
	cleanURI, err := SanitizeRequestURI(uri)
	if err != nil {
		return "", fmt.Errorf("invalid URI: %w", err)
	}

	// Build the URL
	httpsURL := "https://" + cleanHost + cleanURI

	// Final validation
	if err := ValidateRedirectURL(httpsURL); err != nil {
		return "", fmt.Errorf("constructed URL failed validation: %w", err)
	}

	return httpsURL, nil
}
