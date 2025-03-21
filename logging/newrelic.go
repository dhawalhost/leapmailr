package logging

import "github.com/newrelic/go-agent/v3/newrelic"

func AddNewRelic() *newrelic.Application {

	app, err := newrelic.NewApplication(
		newrelic.ConfigAppName("leapmailr"),
		newrelic.ConfigLicense("6e86adefa0544f3f767487b13073dd6cFFFFNRAL"),
		newrelic.ConfigAppLogForwardingEnabled(true),
	)
	if err != nil {
		panic(err)
	}
	return app
}
