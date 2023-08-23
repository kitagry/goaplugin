package calcapi

import (
	"context"

	calc "calc/gen/calc"
	log "calc/gen/log"
)

// calc service example implementation.
// The example methods log the requests and return zero values.
type calcsrvc struct {
	logger *log.Logger
}

// NewCalc returns the calc service implementation.
func NewCalc(logger *log.Logger) calc.Service {
	return &calcsrvc{logger}
}

// Add implements add.
func (s *calcsrvc) Add(ctx context.Context, p *calc.AddPayload) (res int, err error) {
	s.logger.InfoContext(ctx, "calc.add")
	return
}

// Healthz implements healthz.
func (s *calcsrvc) Healthz(ctx context.Context) (err error) {
	s.logger.InfoContext(ctx, "calc.healthz")
	return
}
