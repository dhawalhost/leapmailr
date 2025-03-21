package logging

import (
	"github.com/dhawalhost/leapmailr/config"
	"github.com/newrelic/go-agent/v3/newrelic"
)

func AddNewRelic() *newrelic.Application {
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
