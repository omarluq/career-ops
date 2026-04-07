package scanner

import (
	"context"
	"strings"

	"github.com/chromedp/chromedp"
	"github.com/samber/lo"
	"github.com/samber/mo"
	"github.com/samber/oops"
	"golang.org/x/sync/errgroup"
)

// ScanResult holds the result of scanning a single portal page.
type ScanResult struct {
	URL     string
	Title   string
	Company string
	Portal  string
}

// Scanner manages concurrent portal scanning.
type Scanner struct {
	concurrency int
}

// NewScanner creates a scanner with the given concurrency limit.
func NewScanner(concurrency int) *Scanner {
	if concurrency < 1 {
		concurrency = 1
	}
	return &Scanner{concurrency: concurrency}
}

// ScanPortals scans all configured portals concurrently.
// Returns discovered job listings deduplicated by URL.
func (s *Scanner) ScanPortals(
	ctx context.Context,
	portals []PortalConfig,
	onProgress func(portal string, found int),
) ([]ScanResult, error) {
	// Create a shared Chrome allocator so all tabs share one browser process.
	allocCtx, allocCancel := chromedp.NewExecAllocator(ctx, chromedp.DefaultExecAllocatorOptions[:]...)
	defer allocCancel()

	browserCtx, browserCancel := chromedp.NewContext(allocCtx)
	defer browserCancel()

	// Warm up the browser.
	if err := chromedp.Run(browserCtx); err != nil {
		return nil, oops.Wrapf(err, "starting browser")
	}

	type portalResult struct {
		portal  string
		results []ScanResult
	}

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(s.concurrency)

	ch := make(chan mo.Result[portalResult], len(portals))

	for _, p := range portals {
		g.Go(func() error {
			listings, err := s.scanSinglePortal(browserCtx, gctx, p)
			if err != nil {
				ch <- mo.Err[portalResult](
					oops.Wrapf(err, "scanning portal %s", p.Name),
				)
				return nil
			}

			if onProgress != nil {
				onProgress(p.Name, len(listings))
			}

			ch <- mo.Ok(portalResult{
				portal:  p.Name,
				results: listings,
			})
			return nil
		})
	}

	// Wait for all goroutines, then close channel.
	go func() {
		// Errors are captured via ch, not via g.Wait().
		if waitErr := g.Wait(); waitErr != nil {
			ch <- mo.Err[portalResult](oops.Wrapf(waitErr, "portal scan group"))
		}
		close(ch)
	}()

	var all []ScanResult
	var errs []error

	for r := range ch {
		val, err := r.Get()
		if err != nil {
			errs = append(errs, err)
			continue
		}
		all = append(all, val.results...)
	}

	// Deduplicate by URL.
	all = lo.UniqBy(all, func(r ScanResult) string {
		return strings.ToLower(r.URL)
	})

	if len(errs) > 0 {
		return all, oops.Wrapf(errs[0], "encountered %d portal errors", len(errs))
	}

	return all, nil
}

// scanSinglePortal scans one portal config, which may have a direct URL and/or queries.
func (s *Scanner) scanSinglePortal(
	browserCtx, gctx context.Context,
	p PortalConfig,
) ([]ScanResult, error) {
	var all []ScanResult

	// Scan the base URL if present.
	if p.BaseURL != "" {
		tabCtx, tabCancel := chromedp.NewContext(browserCtx)
		defer tabCancel()

		listings, err := ExtractListings(tabCtx, p.BaseURL)
		if err != nil {
			return nil, oops.Wrapf(err, "scanning base URL %s", p.BaseURL)
		}
		all = append(all, lo.Map(listings, func(r ScanResult, _ int) ScanResult {
			r.Company = p.Name
			r.Portal = p.Name
			return r
		})...)
	}

	// Scan each query URL (for websearch-based portals, these are the query strings).
	// Only scan queries that look like URLs. Search query strings
	// (e.g., site:jobs.ashbyhq.com ...) would need a WebSearch backend
	// which is outside the scope of chromedp extraction.
	urlQueries := lo.Filter(p.Queries, func(q string, _ int) bool {
		return strings.HasPrefix(q, "http")
	})

	lo.ForEach(urlQueries, func(q string, _ int) {
		if gctx.Err() != nil {
			return
		}

		tabCtx, tabCancel := chromedp.NewContext(browserCtx)
		listings, err := ExtractListings(tabCtx, q)
		tabCancel()

		if err != nil {
			return // best-effort per query
		}
		all = append(all, lo.Map(listings, func(r ScanResult, _ int) ScanResult {
			r.Company = p.Name
			r.Portal = p.Name
			return r
		})...)
	})

	return all, nil
}
