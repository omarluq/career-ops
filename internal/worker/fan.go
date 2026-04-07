package worker

import (
	"context"
	"sync"

	"github.com/samber/lo"
	"github.com/samber/mo"
	"github.com/samber/oops"
)

// FanOut distributes items across N workers via channels.
// Results are collected as they complete (unordered).
func FanOut[T any, R any](
	ctx context.Context,
	items <-chan T,
	concurrency int,
	processor func(ctx context.Context, item T) (R, error),
) <-chan mo.Result[R] {
	if concurrency < 1 {
		concurrency = 1
	}

	out := make(chan mo.Result[R])
	var wg sync.WaitGroup

	lo.Times(concurrency, func(_ int) struct{} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for item := range items {
				if ctx.Err() != nil {
					out <- mo.Err[R](oops.Wrapf(ctx.Err(), "context canceled during fan-out"))
					return
				}

				r, err := processor(ctx, item)
				if err != nil {
					out <- mo.Err[R](oops.Wrapf(err, "fan-out processor"))
				} else {
					out <- mo.Ok(r)
				}
			}
		}()
		return struct{}{}
	})

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}
