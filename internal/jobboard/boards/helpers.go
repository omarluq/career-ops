// Package boards provides concrete job board implementations for the jobboard
// package. Each board implements the jobboard.Board interface with platform-specific
// search and apply logic.
package boards

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/samber/lo"
	"github.com/samber/oops"

	"github.com/omarluq/career-ops/internal/jobboard"
)

const remoteQueryParam = "&remote=true"

// scrapeJobLinks navigates to a URL with chromedp and extracts all anchor elements
// whose href matches the given linkPattern substring. Results are capped at limit entries.
func scrapeJobLinks(
	ctx context.Context, pageURL, linkPattern, boardSlug string, limit int,
) ([]jobboard.SearchResult, error) {
	allocCtx, allocCancel := chromedp.NewExecAllocator(ctx, chromedp.DefaultExecAllocatorOptions[:]...)
	defer allocCancel()

	taskCtx, taskCancel := chromedp.NewContext(allocCtx)
	defer taskCancel()

	var hrefs, titles []string

	err := chromedp.Run(taskCtx,
		chromedp.Navigate(pageURL),
		chromedp.WaitReady("body"),
		chromedp.EvaluateAsDevTools(`
			Array.from(document.querySelectorAll('a[href]'))
				.filter(a => a.href.includes('`+linkPattern+`'))
				.map(a => a.href)
		`, &hrefs),
		chromedp.EvaluateAsDevTools(`
			Array.from(document.querySelectorAll('a[href]'))
				.filter(a => a.href.includes('`+linkPattern+`'))
				.map(a => (a.textContent || '').trim())
		`, &titles),
	)
	if err != nil {
		return nil, oops.
			In("boards").
			Tags("scrape", boardSlug).
			Wrapf(err, "scraping job links from %s", pageURL)
	}

	count := lo.Min([]int{len(hrefs), len(titles), limit})

	results := lo.Map(lo.Range(count), func(i int, _ int) jobboard.SearchResult {
		return jobboard.SearchResult{
			URL:      hrefs[i],
			Title:    titles[i],
			Board:    boardSlug,
			PostedAt: time.Time{},
		}
	})

	return results, nil
}

// httpGet performs a GET request to the given URL and returns the response body bytes.
func httpGet(ctx context.Context, url string) (retBody []byte, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, oops.Wrapf(err, "creating GET request for %s", url)
	}

	req.Header.Set("User-Agent", "career-ops/1.0")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, oops.Wrapf(err, "executing GET request for %s", url)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	if resp.StatusCode >= 400 {
		return nil, oops.Errorf("HTTP %d from %s", resp.StatusCode, url)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, oops.Wrapf(err, "reading response body from %s", url)
	}

	return body, nil
}

// buildSearchURL constructs a search URL by appending keyword query parameters.
func buildSearchURL(base string, keywords []string) string {
	if len(keywords) == 0 {
		return base
	}

	sep := "?"
	if strings.Contains(base, "?") {
		sep = "&"
	}

	return base + sep + "q=" + strings.Join(keywords, "+")
}

// defaultMax returns the effective result limit, falling back to 50 if unset.
func defaultMax(n int) int {
	if n <= 0 {
		return 50
	}

	return n
}

// healthCheck verifies a board is reachable by issuing a HEAD request to the target URL.
func healthCheck(ctx context.Context, targetURL string) (err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, targetURL, http.NoBody)
	if err != nil {
		return oops.Wrapf(err, "creating health check request for %s", targetURL)
	}

	req.Header.Set("User-Agent", "career-ops/1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return oops.Wrapf(err, "health check failed for %s", targetURL)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	if resp.StatusCode >= 500 {
		return oops.Errorf("health check returned HTTP %d for %s", resp.StatusCode, targetURL)
	}

	return nil
}

// applyNotSupported is a convenience Apply implementation for search-only boards.
func applyNotSupported(_ context.Context, _ jobboard.Application) (jobboard.ApplyResult, error) {
	return jobboard.ApplyResult{}, jobboard.ErrApplyNotSupported
}
