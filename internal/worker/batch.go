package worker

import (
	"context"
	"sync/atomic"

	"github.com/samber/mo"
)

// BatchConfig configures a batch processing run.
type BatchConfig[T any, R any] struct {
	Processor   func(ctx context.Context, item T) (R, error)
	OnProgress  func(completed, total int)
	Items       []T
	Concurrency int
}

// RunBatch executes a batch of work with progress reporting.
func RunBatch[T any, R any](ctx context.Context, cfg BatchConfig[T, R]) ([]mo.Result[R], error) {
	total := len(cfg.Items)
	if total == 0 {
		return nil, nil
	}

	var completed atomic.Int64

	wrapped := func(ctx context.Context, item T) (R, error) {
		r, err := cfg.Processor(ctx, item)

		n := int(completed.Add(1))
		if cfg.OnProgress != nil {
			cfg.OnProgress(n, total)
		}

		return r, err
	}

	pool := NewPool[T, R](cfg.Concurrency, wrapped)
	return pool.Run(ctx, cfg.Items)
}
