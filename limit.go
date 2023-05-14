package csv2opensearch

import (
	"context"

	"golang.org/x/time/rate"
)

// RateLimiter is a thin wrapper over rate.Limiter that throttles the sink ingestion rate.
type RateLimiter struct {
	limiter *rate.Limiter
	sink    Sink
}

// NewRateLimiter configures the internal rate limiter and connects it to the sink.
func NewRateLimiter(r rate.Limit, burst int, sink Sink) *RateLimiter {
	return &RateLimiter{limiter: rate.NewLimiter(r, burst), sink: sink}
}

// Write throttles writes to the sink based on the current rate.
func (rl *RateLimiter) Write(ctx context.Context, v []string) error {
	rl.limiter.WaitN(ctx, len(v))
	return rl.sink.Write(ctx, v)
}
