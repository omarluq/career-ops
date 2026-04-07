package worker_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/omarluq/career-ops/internal/worker"
	"github.com/samber/lo"
	"github.com/samber/mo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPool_Success(t *testing.T) {
	t.Parallel()

	pool := worker.NewPool[int, int](3, func(_ context.Context, item int) (int, error) {
		return item * 2, nil
	})

	results, err := pool.Run(context.Background(), []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
	require.NoError(t, err)
	require.Len(t, results, 10)

	lo.ForEach(results, func(r mo.Result[int], i int) {
		val, err := r.Get()
		require.NoError(t, err)
		assert.Equal(t, (i+1)*2, val)
	})
}

func TestPool_PartialFailure(t *testing.T) {
	t.Parallel()

	errBoom := errors.New("boom")

	pool := worker.NewPool[int, int](3, func(_ context.Context, item int) (int, error) {
		if item%3 == 0 {
			return 0, errBoom
		}
		return item * 2, nil
	})

	results, err := pool.Run(context.Background(), []int{1, 2, 3, 4, 5, 6})
	require.Error(t, err)
	require.Len(t, results, 6)

	// Items 1,2,4,5 succeed; items 3,6 fail.
	lo.ForEach(results, func(r mo.Result[int], i int) {
		item := i + 1
		if item%3 == 0 {
			_, rErr := r.Get()
			assert.Error(t, rErr)
		} else {
			val, rErr := r.Get()
			assert.NoError(t, rErr)
			assert.Equal(t, item*2, val)
		}
	})
}

func TestPool_ContextCancel(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	var started atomic.Int64

	pool := worker.NewPool[int, int](2, func(ctx context.Context, item int) (int, error) {
		started.Add(1)
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-time.After(200 * time.Millisecond):
			return item, nil
		}
	})

	// Cancel after a short delay so not all items complete.
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	results, runErr := pool.Run(ctx, []int{1, 2, 3, 4, 5, 6, 7, 8})
	// Cancellation may or may not cause a pool-level error depending on timing.
	if runErr != nil {
		t.Logf("pool returned error (expected during cancellation): %v", runErr)
	}
	require.Len(t, results, 8)

	// At least some results should be errors due to cancellation.
	errCount := lo.CountBy(results, func(r mo.Result[int]) bool {
		_, err := r.Get()
		return err != nil
	})
	assert.Greater(t, errCount, 0, "expected at least one canceled result")
}

func TestPool_OrderPreserved(t *testing.T) {
	t.Parallel()

	pool := worker.NewPool[int, string](4, func(_ context.Context, item int) (string, error) {
		// Varying sleep to exercise ordering under concurrency.
		time.Sleep(time.Duration(10-item) * time.Millisecond)
		letters := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J"}
		return letters[item], nil
	})

	items := []int{0, 1, 2, 3, 4}
	results, err := pool.Run(context.Background(), items)
	require.NoError(t, err)
	require.Len(t, results, 5)

	expected := []string{"A", "B", "C", "D", "E"}
	lo.ForEach(results, func(r mo.Result[string], i int) {
		val, rErr := r.Get()
		require.NoError(t, rErr)
		assert.Equal(t, expected[i], val)
	})
}

func TestFanOut_Basic(t *testing.T) {
	t.Parallel()

	in := make(chan int, 5)
	lo.ForEach(lo.RangeFrom(1, 5), func(i int, _ int) {
		in <- i
	})
	close(in)

	out := worker.FanOut[int, int](
		context.Background(),
		in,
		2,
		func(_ context.Context, item int) (int, error) {
			return item * 10, nil
		},
	)

	collected := lo.ChannelToSlice(out)
	vals := lo.Map(collected, func(r mo.Result[int], _ int) int {
		val, err := r.Get()
		require.NoError(t, err)
		return val
	})

	assert.Len(t, vals, 5)
	// Unordered, so just check all expected values are present.
	assert.ElementsMatch(t, []int{10, 20, 30, 40, 50}, vals)
}
