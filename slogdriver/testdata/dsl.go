package testdata

import (
	. "goa.design/goa/v3/dsl"

	slogdriver "github.com/kitagry/goaplugin/slogdriver/dsl"
)

var SimpleServiceDSL = func() {
	slogdriver.HealthCheckPaths("/liveness", "/readiness", "/healthz")

	Service("SimpleService", func() {
		Method("SimpleMethod", func() {
			HTTP(func() {
				GET("/")
			})
		})
	})
}
