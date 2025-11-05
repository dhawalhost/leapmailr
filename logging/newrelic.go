package logging

import (
	"fmt"

	"github.com/dhawalhost/leapmailr/config"
	"github.com/newrelic/go-agent/v3/newrelic"
	"go.uber.org/zap"
)

// InitNewRelic initializes New Relic APM if license key is configured
// Returns nil if license key is not set (allowing app to run without New Relic)
func InitNewRelic(logger *zap.Logger) *newrelic.Application {
	var conf = config.GetConfig()

	// If no license key, log warning and return nil
	if conf.NRLicense == "" {
		logger.Warn("New Relic license key not configured - APM monitoring disabled",
			zap.String("hint", "Set NR_LICENSE_KEY environment variable to enable New Relic"))
		return nil
	}

	logger.Info("Initializing New Relic APM",
		zap.String("app_name", "leapmailr"),
		zap.Bool("log_forwarding", true))

	app, err := newrelic.NewApplication(
		newrelic.ConfigAppName("leapmailr"),
		newrelic.ConfigLicense(conf.NRLicense),
		newrelic.ConfigAppLogForwardingEnabled(true),
		newrelic.ConfigDistributedTracerEnabled(true),
	)

	if err != nil {
		logger.Error("Failed to initialize New Relic APM",
			zap.Error(err),
			zap.String("hint", "Check your NR_LICENSE_KEY value"))
		return nil
	}

	logger.Info("New Relic APM initialized successfully")
	return app
}

// Deprecated: Use InitNewRelic instead
func AddNewRelic() *newrelic.Application {
	fmt.Println("Warning: AddNewRelic is deprecated, use InitNewRelic instead")
	var conf = config.GetConfig()
	if conf.NRLicense == "" {
		panic("license not found")
	}
	app, err := newrelic.NewApplication(
		newrelic.ConfigAppName("leapmailr"),
		newrelic.ConfigLicense(conf.NRLicense),
		newrelic.ConfigAppLogForwardingEnabled(true),
	)
	if err != nil {
		panic(err)
	}
	return app
}
