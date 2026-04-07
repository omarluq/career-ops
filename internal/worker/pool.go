// Package worker provides generic concurrency primitives for bounded parallel processing.
package worker

import (
	"context"
	"sync"

	"github.com/samber/mo"
	"github.com/samber/oops"
	"golang.org/x/sync/errgroup"
)

// Pool manages a bounded set of concurrent workers.
type Pool[T any, R any] struct {
	processor   func(ctx context.Context, item T) (R, error)
	concurrency int
}

// NewPool creates a worker pool with the given concurrency limit and processor function.
func NewPool[T any, R any](
	concurrency int,
	processor func(ctx context.Context, item T) (R, error),
) *Pool[T, R] {
	if concurrency < 1 {
		concurrency = 1
	}
	return &Pool[T, R]{
		concurrency: concurrency,
		processor:   processor,
	}
}

// Run processes all items concurrently (bounded by concurrency limit).
// Returns results in order and the first error encountered (if any).
func (p *Pool[T, R]) Run(ctx context.Context, items []T) ([]mo.Result[R], error) {
	n := len(items)
	if n == 0 {
		return nil, nil
	}

	results := make([]mo.Result[R], n)
	var mu sync.Mutex
	var firstErr error

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(p.concurrency)

	for i, item := range items {
		i, item := i, item
		g.Go(func() error {
			if ctxErr := gctx.Err(); ctxErr != nil {
				results[i] = mo.Err[R](oops.Wrapf(ctxErr, "context canceled before processing item %d", i))
				return nil
			}

			r, err := p.processor(gctx, item)
			if err != nil {
				results[i] = mo.Err[R](oops.Wrapf(err, "processing item %d", i))
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
				return nil
			}

			results[i] = mo.Ok(r)
			return nil
		})
	}

	// errgroup goroutines always return nil so Wait never returns an error itself.
	// Errors are captured via firstErr, not the errgroup return.
	if waitErr := g.Wait(); waitErr != nil {
		return results, oops.Wrapf(waitErr, "pool wait")
	}

	return results, firstErr
}
