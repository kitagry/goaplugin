package expr

import (
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

var Root = &RootExpr{
	HealthChecks: []*HealthCheckExpr{},
}

type RootExpr struct {
	HealthChecks []*HealthCheckExpr
}

func init() {
	eval.Register(Root)
}

// EvalName returns the name used in error messages.
func (r *RootExpr) EvalName() string {
	return "slogdriver plugin"
}

// WalkSets iterates over the API-level and service-level CORS definitions.
func (r *RootExpr) WalkSets(walk eval.SetWalker) {
	hexps := make(eval.ExpressionSet, 0, len(r.HealthChecks))
	for _, o := range r.HealthChecks {
		hexps = append(hexps, o)
	}
	walk(hexps)
}

// DependsOn tells the eval engine to run the goa DSL first.
func (r *RootExpr) DependsOn() []eval.Root {
	return []eval.Root{expr.Root}
}

// Packages returns the import path to the Go packages that make
// up the DSL. This is used to skip frames that point to files
// in these packages when computing the location of errors.
func (r *RootExpr) Packages() []string {
	return []string{"github.com/kitagry/goaplugin/slogdriver/dsl"}
}
