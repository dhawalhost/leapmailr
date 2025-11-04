package main

import (
	"fmt"
	"log"

	"github.com/dhawalhost/leapmailr/config"
	"github.com/dhawalhost/leapmailr/database"
	"github.com/dhawalhost/leapmailr/handlers"
	"github.com/dhawalhost/leapmailr/logging"
	"github.com/dhawalhost/leapmailr/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	conf := config.LoadConfig()
	if conf.EnvMode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize logger
	logger := logging.InitLogger()
	defer func() {
		if err := logger.Sync(); err != nil {
			fmt.Println(err)
		}
	}()
	if conf.EnvMode == "release" {
		logger.Info("release mode")
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

	defer func() {
		fmt.Println("Closing DB!!!")
		if err := database.CloseDB(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	// Setup Gin router
	r := gin.New()
	r.Use(middleware.CorsMiddleware())
	r.Use(middleware.LoggerMiddleware())
	r.Use(gin.Recovery())

	// Public routes
	r.GET("/health", handlers.HandleHealthCheck)

	// API v1 routes
	api := r.Group("/api/v1")
	{
		// Authentication routes (public)
		auth := api.Group("/auth")
		{
			auth.POST("/register", handlers.RegisterHandler)
			auth.POST("/login", handlers.LoginHandler)
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
