package applicator

import (
	"context"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/samber/lo"
	"github.com/samber/mo"
	"github.com/samber/oops"
	"golang.org/x/sync/errgroup"

	"github.com/omarluq/career-ops/internal/db"
)

// Applicator orchestrates high-throughput application submissions with
// per-portal rate limiting and ATS-specific form handlers.
type Applicator struct {
	repo        db.Repository
	limiter     *RateLimiter
	concurrency int
}

// New creates an Applicator with the given repository, concurrency limit,
// and per-portal rate limit interval.
func New(r db.Repository, concurrency int, rateLimit time.Duration) *Applicator {
	if concurrency < 1 {
		concurrency = 1
	}

	return &Applicator{
		repo:        r,
		concurrency: concurrency,
		limiter:     NewRateLimiter(rateLimit),
	}
}

// SubmitBatch submits applications for a batch of evaluated jobs.
// Each job must have been evaluated (score > threshold) and have a generated CV PDF.
// The onProgress callback is invoked after each submission completes (may be nil).
// Returns ordered results corresponding to the input jobs slice.
func (a *Applicator) SubmitBatch(
	ctx context.Context,
	jobs []SubmissionJob,
	onProgress func(job SubmissionJob, result SubmissionResult),
) ([]SubmissionResult, error) {
	if len(jobs) == 0 {
		return nil, nil
	}

	// Shared browser allocator so all tabs reuse one Chrome process.
	allocCtx, allocCancel := chromedp.NewExecAllocator(
		ctx, chromedp.DefaultExecAllocatorOptions[:]...,
	)
	defer allocCancel()

	browserCtx, browserCancel := chromedp.NewContext(allocCtx)
	defer browserCancel()

	// Warm up the browser before launching concurrent work.
	if err := chromedp.Run(browserCtx); err != nil {
		return nil, oops.Wrapf(err, "starting browser for batch submission")
	}

	results := make([]mo.Result[SubmissionResult], len(jobs))

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(a.concurrency)

	lo.ForEach(jobs, func(job SubmissionJob, i int) {
		g.Go(func() error {
			res := a.submitOne(gctx, browserCtx, job)

			results[i] = mo.Ok(res)

			if onProgress != nil {
				onProgress(job, res)
			}

			return nil
		})
	})

	if waitErr := g.Wait(); waitErr != nil {
		return nil, oops.Wrapf(waitErr, "batch submission wait")
	}

	out := lo.Map(results, func(r mo.Result[SubmissionResult], _ int) SubmissionResult {
		val, err := r.Get()
		if err != nil {
			return SubmissionResult{
				Status:       StatusFailed,
				ErrorMessage: err.Error(),
				SubmittedAt:  time.Now(),
			}
		}
		return val
	})

	return out, nil
}

// submitOne handles rate limiting, handler dispatch, and error capture
// for a single submission job.
func (a *Applicator) submitOne(
	ctx context.Context,
	browserCtx context.Context,
	job SubmissionJob,
) SubmissionResult {
	// Enforce per-portal rate limit.
	portal := job.Portal
	if portal == "" {
		portal = DetectPortal(job.JDURL)
		job.Portal = portal
	}

	if err := a.limiter.Wait(ctx, portal); err != nil {
		return SubmissionResult{
			Job:          job,
			Status:       StatusFailed,
			ErrorMessage: oops.Wrapf(err, "rate limit wait").Error(),
			SubmittedAt:  time.Now(),
		}
	}

	handler := HandlerFor(job.JDURL)

	result, err := handler.Submit(ctx, browserCtx, job)
	if err != nil {
		return SubmissionResult{
			Job:          job,
			Status:       StatusFailed,
			ErrorMessage: oops.Wrapf(err, "submitting to %s", portal).Error(),
			SubmittedAt:  time.Now(),
		}
	}

	return result
}
