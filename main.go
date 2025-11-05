package main

import (
	"fmt"
	"log"
	"time"

	"github.com/dhawalhost/leapmailr/config"
	"github.com/dhawalhost/leapmailr/database"
	"github.com/dhawalhost/leapmailr/handlers"
	"github.com/dhawalhost/leapmailr/logging"
	"github.com/dhawalhost/leapmailr/middleware"
	"github.com/dhawalhost/leapmailr/monitoring"
	"github.com/dhawalhost/leapmailr/utils"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	conf := config.LoadConfig()
	if conf.EnvMode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize custom validator with enhanced rules (GAP-SEC-013)
	customValidator := utils.NewCustomValidator()
	binding.Validator = &validatorAdapter{validator: customValidator.Validator}

	// Initialize monitoring metrics (GAP-AV-003)
	monitoring.InitMetrics("1.0.0", conf.EnvMode)

	// Update uptime metric every minute
	go func() {
		startTime := time.Now()
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			monitoring.AppUptimeSeconds.Set(time.Since(startTime).Seconds())
		}
	}()

	// Initialize logger (GAP-SEC-009: Centralized structured logging)
	logger := logging.InitLogger()
	defer func() {
		if err := logger.Sync(); err != nil {
			fmt.Println(err)
		}
	}()

	// Log startup information
	logger.Info("Starting LeapMailR API Server",
		zap.String("version", "1.0.0"),
		zap.String("environment", conf.EnvMode),
		zap.String("port", conf.Port),
	)

	if conf.EnvMode == "release" {
		logger.Info("release mode")
	}

	// Initialize New Relic APM (optional - runs without it if not configured)
	nrApp := logging.InitNewRelic(logger)
	if nrApp != nil {
		logger.Info("New Relic APM integration enabled")
	}

	// Initialize database
	if err := database.InitDatabase(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	fmt.Println("Database initialized successfully")

	// Seed default templates
	if err := database.SeedDefaultTemplates(); err != nil {
		log.Printf("Warning: Failed to seed default templates: %v", err)
	} else {
		fmt.Println("Default templates seeded successfully")
	}

	// Initialize health checks (GAP-AV-003)
	handlers.InitializeHealthChecks()
	fmt.Println("Health checks initialized")

	// Initialize MFA service (GAP-SEC-001)
	encryption, err := utils.NewEncryptionService()
	if err != nil {
		log.Fatalf("Failed to initialize encryption service: %v", err)
	}
	handlers.InitMFAService(encryption)
	fmt.Println("MFA service initialized")

	defer func() {
		fmt.Println("Closing DB!!!")
		if err := database.CloseDB(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	// Setup Gin router
	r := gin.New()

	// New Relic middleware (GAP-AV-003) - first in chain for full transaction tracking
	r.Use(middleware.NewRelicMiddleware(nrApp))

	// Metrics middleware (GAP-AV-003) - early in chain to measure all requests
	r.Use(middleware.PrometheusMetrics())

	// Structured logging middleware (GAP-SEC-009) - adds correlation IDs
	r.Use(middleware.StructuredLogger())

	// Security middlewares (GAP-SEC-011, GAP-SEC-012)
	r.Use(middleware.RedirectToHTTPS(conf.EnvMode)) // Redirect HTTP to HTTPS
	r.Use(middleware.SecurityHeaders())             // Add security headers (HSTS, CSP, etc.)
	r.Use(middleware.CorsMiddleware(conf.EnvMode))  // Strict CORS with whitelist
	r.Use(middleware.TrustedProxyHeaders())         // Handle proxy headers correctly
	r.Use(gin.Recovery())                           // Panic recovery
	r.Use(middleware.EnhancedRateLimiter(nil))      // Multi-tier rate limiting (GAP-SEC-010)

	// Input validation and sanitization middlewares (GAP-SEC-013)
	r.Use(middleware.ContentTypeValidator())     // Validate content-type headers
	r.Use(middleware.InputSanitizer())           // Sanitize all inputs
	r.Use(middleware.XSSProtection())            // Add XSS protection headers
	r.Use(middleware.ValidateEmailAttachments()) // Validate email attachments

	// CSRF protection middleware (GAP-SEC-014) - after auth, validates X-CSRF-Token header
	r.Use(middleware.CSRFProtection())

	// Request size limit (10MB for file uploads, adjust as needed)
	r.Use(middleware.RequestSizeLimit(10 * 1024 * 1024))

	// Public health and metrics routes (GAP-AV-003)
	r.GET("/health", handlers.HandleHealthCheck)          // Detailed health check
	r.GET("/health/ready", handlers.HandleReadinessCheck) // Kubernetes readiness probe
	r.GET("/health/live", handlers.HandleLivenessCheck)   // Kubernetes liveness probe
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))      // Prometheus metrics endpoint

	// Email tracking routes (public - no auth required for pixel and click tracking)
	r.GET("/api/v1/track/open/:pixel_id", handlers.TrackOpenHandler)
	r.GET("/api/v1/track/click/:pixel_id/:link_id", handlers.TrackClickHandler)

	// API v1 routes
	api := r.Group("/api/v1")
	{
		// Authentication routes (public)
		auth := api.Group("/auth")
		{
			auth.POST("/register", handlers.RegisterHandler)
			auth.POST("/login", handlers.LoginHandler)
			auth.POST("/login-mfa", handlers.LoginWithMFAHandler)                // MFA login (GAP-SEC-001)
			auth.POST("/login-backup-code", handlers.LoginWithBackupCodeHandler) // Backup code login (GAP-SEC-001)
			auth.POST("/refresh", handlers.RefreshTokenHandler)
		}

		// EmailJS-compatible routes (with authentication)
		api.POST("/email/send", handlers.AuthMiddleware(), handlers.SendEmailHandler)
		api.POST("/email/send-form", handlers.SendFormHandler) // EmailJS compatibility - uses API key
		api.POST("/email/send-bulk", handlers.AuthMiddleware(), handlers.SendBulkEmailHandler)

		// Webhook routes (public - will be secured by provider-specific tokens/signatures)
		webhooks := api.Group("/webhooks")
		{
			webhooks.POST("/:provider", handlers.WebhookHandler)      // SendGrid, Mailgun, etc.
			webhooks.POST("/generic", handlers.GenericWebhookHandler) // Custom webhooks
		}

		// Protected routes (require authentication)
		protected := api.Group("/", handlers.AuthMiddleware())
		{
			// User management
			protected.GET("/profile", handlers.ProfileHandler)
			protected.POST("/auth/logout", handlers.LogoutHandler)

			// Project management
			projects := protected.Group("/projects")
			{
				projects.GET("/default", handlers.GetDefaultProject) // Must come before /:id
				projects.POST("", handlers.CreateProject)
				projects.GET("", handlers.GetProjects)
				projects.GET("/:id", handlers.GetProject)
				projects.PUT("/:id", handlers.UpdateProject)
				projects.DELETE("/:id", handlers.DeleteProject)
				projects.POST("/:id/default", handlers.SetDefaultProject)
			}

			// Email management
			protected.GET("/emails", handlers.GetEmailHistoryHandler)
			protected.GET("/emails/:id", handlers.GetEmailStatusHandler)

			// Template management
			templates := protected.Group("/templates")
			{
				templates.GET("/defaults", handlers.GetDefaultTemplatesHandler)     // Get default templates
				templates.GET("/categories", handlers.GetTemplateCategoriesHandler) // Get categories
				templates.POST("", handlers.CreateTemplateHandler)
				templates.GET("", handlers.ListTemplatesHandler)
				templates.GET("/:id", handlers.GetTemplateHandler)
				templates.PUT("/:id", handlers.UpdateTemplateHandler)
				templates.DELETE("/:id", handlers.DeleteTemplateHandler)
				templates.POST("/:id/test", handlers.TestTemplateHandler)
				templates.POST("/:id/clone", handlers.CloneTemplateHandler)
				templates.GET("/:id/versions", handlers.GetTemplateVersionsHandler)
			}

			// Email service management
			services := protected.Group("/email-services")
			{
				services.GET("/providers", handlers.GetSMTPProvidersHandler)                     // Get all SMTP providers
				services.GET("/providers/:id", handlers.GetSMTPProviderHandler)                  // Get specific provider
				services.GET("/providers-categories", handlers.GetSMTPProviderCategoriesHandler) // Get categories
				services.POST("", handlers.CreateEmailServiceHandler)
				services.GET("", handlers.ListEmailServicesHandler)
				services.GET("/:id", handlers.GetEmailServiceHandler)
				services.GET("/:id/config", handlers.GetEmailServiceConfigHandler) // Debug: view config
				services.PUT("/:id", handlers.UpdateEmailServiceHandler)
				services.DELETE("/:id", handlers.DeleteEmailServiceHandler)
				services.POST("/:id/test", handlers.TestEmailServiceHandler)
				services.POST("/:id/default", handlers.SetDefaultServiceHandler)
			}

			// CAPTCHA configuration management
			captcha := protected.Group("/captcha")
			{
				captcha.POST("", handlers.CreateCaptchaConfigHandler)
				captcha.GET("", handlers.ListCaptchaConfigsHandler)
				captcha.GET("/:id", handlers.GetCaptchaConfigHandler)
				captcha.PUT("/:id", handlers.UpdateCaptchaConfigHandler)
				captcha.DELETE("/:id", handlers.DeleteCaptchaConfigHandler)
			}

			// Suppression list management
			suppressions := protected.Group("/suppressions")
			{
				suppressions.POST("", handlers.AddSuppressionHandler)
				suppressions.POST("/bulk", handlers.AddBulkSuppressionsHandler)
				suppressions.GET("", handlers.ListSuppressionsHandler)
				suppressions.GET("/check", handlers.CheckSuppressionHandler)
				suppressions.DELETE("/:id", handlers.DeleteSuppressionHandler)
			}

			// Auto-reply configuration management
			autoreplies := protected.Group("/autoreplies")
			{
				autoreplies.POST("", handlers.CreateAutoReplyHandler)
				autoreplies.GET("", handlers.ListAutoRepliesHandler)
				autoreplies.GET("/:id", handlers.GetAutoReplyHandler)
				autoreplies.PUT("/:id", handlers.UpdateAutoReplyHandler)
				autoreplies.DELETE("/:id", handlers.DeleteAutoReplyHandler)
				autoreplies.POST("/:id/test", handlers.TestAutoReplyHandler)
				autoreplies.GET("/logs", handlers.GetAutoReplyLogsHandler)
			}

			// Email tracking and analytics
			tracking := protected.Group("/analytics")
			{
				tracking.GET("/email/:email_id", handlers.GetEmailAnalyticsHandler)
				tracking.GET("/email/:email_id/events", handlers.GetEmailTrackingEventsHandler)
				tracking.GET("/campaign/:campaign_id", handlers.GetCampaignAnalyticsHandler)
			}

			// API Key Pair management (for SDK usage)
			apikeys := protected.Group("/api-keys")
			{
				apikeys.POST("", handlers.GenerateAPIKeyPairHandler)
				apikeys.GET("", handlers.ListAPIKeyPairsHandler)
				apikeys.GET("/:id", handlers.GetAPIKeyPairHandler)
				apikeys.PUT("/:id", handlers.UpdateAPIKeyPairHandler)
				apikeys.POST("/:id/revoke", handlers.RevokeAPIKeyPairHandler)
				apikeys.DELETE("/:id", handlers.DeleteAPIKeyPairHandler)
				apikeys.POST("/:id/rotate", handlers.RotateAPIKeyPairHandler)
				apikeys.GET("/:id/usage", handlers.GetAPIKeyUsageHandler)
			}

			// Contact management
			contacts := protected.Group("/contacts")
			{
				contacts.POST("", handlers.CreateContactHandler)
				contacts.GET("", handlers.ListContactsHandler)
				contacts.GET("/stats", handlers.GetContactStatsHandler)
				contacts.GET("/export", handlers.ExportContactsHandler)
				contacts.GET("/:id", handlers.GetContactHandler)
				contacts.PUT("/:id", handlers.UpdateContactHandler)
				contacts.DELETE("/:id", handlers.DeleteContactHandler)
				contacts.POST("/import", handlers.ImportContactsHandler)
			}

			// Multi-Factor Authentication (GAP-SEC-001)
			mfa := protected.Group("/mfa")
			{
				mfa.POST("/setup", handlers.SetupMFAHandler)                                // Start MFA setup
				mfa.POST("/verify-setup", handlers.VerifyMFASetupHandler)                   // Complete MFA setup
				mfa.POST("/disable", handlers.DisableMFAHandler)                            // Disable MFA
				mfa.POST("/regenerate-backup-codes", handlers.RegenerateBackupCodesHandler) // Regenerate backup codes
				mfa.GET("/status", handlers.GetMFAStatusHandler)                            // Get MFA status
			}

			// TODO: Add more protected routes for:
			// - Analytics
			// - Organization management
		}
	}

	// Start server
	log.Printf("Server starting on port %s", conf.Port)
	if err := r.Run(fmt.Sprintf(":%v", conf.Port)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

}

// validatorAdapter adapts our custom validator to Gin's validator interface
type validatorAdapter struct {
	validator *validator.Validate
}

func (v *validatorAdapter) ValidateStruct(obj interface{}) error {
	return v.validator.Struct(obj)
}

func (v *validatorAdapter) Engine() interface{} {
	return v.validator
}
