package expr

type HealthCheckExpr struct {
	Paths []string
}

// EvalName returns the generic expression name used in error messages.
func (o *HealthCheckExpr) EvalName() string {
	return "HealthCheck"
}
