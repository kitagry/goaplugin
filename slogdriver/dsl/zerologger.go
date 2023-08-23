package dsl

import (
	"github.com/kitagry/goaplugin/slogdriver/expr"

	// Register code generators for the slogdriver plugin
	_ "github.com/kitagry/goaplugin/slogdriver"
)

func HealthCheckPaths(paths ...string) {
	hexpr := &expr.HealthCheckExpr{Paths: paths}
	expr.Root.HealthChecks = append(expr.Root.HealthChecks, hexpr)
}
