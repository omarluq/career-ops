package scanner

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/chromedp/chromedp"
	"github.com/samber/lo"
	"github.com/samber/oops"
)

// BoardAdapter defines how to discover jobs from a specific board.
type BoardAdapter interface {
	Name() string
	DiscoverJobs(
		ctx context.Context,
		browserCtx context.Context,
		config BoardConfig,
	) ([]ScanResult, error)
}

// BoardConfig holds search parameters for board adapters.
type BoardConfig struct {
	Location   string
	Keywords   []string
	MaxResults int
	Remote     bool
}

// queryString joins keywords into a single search string.
func (bc BoardConfig) queryString() string {
	return strings.Join(bc.Keywords, " ")
}

// --- LinkedIn Adapter ---

// LinkedInAdapter scans LinkedIn public job search.
// Note: LinkedIn limits results without login; MVP uses the public endpoint.
type LinkedInAdapter struct{}

// Name returns the adapter display name.
func (a *LinkedInAdapter) Name() string { return "LinkedIn" }

// DiscoverJobs searches LinkedIn jobs using the public search URL.
func (a *LinkedInAdapter) DiscoverJobs(
	_ context.Context,
	browserCtx context.Context,
	config BoardConfig,
) ([]ScanResult, error) {
	params := url.Values{
		"keywords": {config.queryString()},
	}
	if config.Location != "" {
		params.Set("location", config.Location)
	}
	if config.Remote {
		params.Set("f_WT", "2") // remote filter
	}

	searchURL := "https://www.linkedin.com/jobs/search/?" + params.Encode()

	return discoverFromURL(browserCtx, searchURL, "linkedin.com/jobs/view/", a.Name(), config.MaxResults)
}

// --- YCombinator Adapter ---

// YCombinatorAdapter scans Work at a Startup (workatastartup.com).
type YCombinatorAdapter struct{}

// Name returns the adapter display name.
func (a *YCombinatorAdapter) Name() string { return "YCombinator" }

// DiscoverJobs searches YC Work at a Startup.
func (a *YCombinatorAdapter) DiscoverJobs(
	_ context.Context,
	browserCtx context.Context,
	config BoardConfig,
) ([]ScanResult, error) {
	params := queryParamsWithRemote(config)
	searchURL := "https://www.workatastartup.com/jobs?" + params.Encode()

	return discoverFromURL(browserCtx, searchURL, "/jobs/", a.Name(), config.MaxResults)
}

// --- Wellfound Adapter ---

// WellfoundAdapter scans Wellfound (formerly AngelList Talent).
type WellfoundAdapter struct{}

// Name returns the adapter display name.
func (a *WellfoundAdapter) Name() string { return "Wellfound" }

// DiscoverJobs searches Wellfound job listings.
func (a *WellfoundAdapter) DiscoverJobs(
	_ context.Context,
	browserCtx context.Context,
	config BoardConfig,
) ([]ScanResult, error) {
	params := queryParamsWithRemote(config)
	searchURL := "https://wellfound.com/jobs?" + params.Encode()

	return discoverFromURL(browserCtx, searchURL, "/jobs/", a.Name(), config.MaxResults)
}

// --- AI Jobs Adapter ---

// AIJobsAdapter aggregates results from AI-focused job boards.
type AIJobsAdapter struct{}

// Name returns the adapter display name.
func (a *AIJobsAdapter) Name() string { return "AIJobs" }

// DiscoverJobs searches AI-specific job boards.
func (a *AIJobsAdapter) DiscoverJobs(
	ctx context.Context,
	browserCtx context.Context,
	config BoardConfig,
) ([]ScanResult, error) {
	type boardSource struct {
		name    string
		urlTpl  string
		pattern string
	}

	sources := []boardSource{
		{
			name:    "ai-jobs.net",
			urlTpl:  "https://ai-jobs.net/search/?q=%s",
			pattern: "ai-jobs.net/job/",
		},
	}

	var all []ScanResult

	lo.ForEach(sources, func(src boardSource, _ int) {
		if ctx.Err() != nil {
			return
		}

		searchURL := fmt.Sprintf(src.urlTpl, url.QueryEscape(config.queryString()))
		portalName := fmt.Sprintf("%s/%s", a.Name(), src.name)

		results, err := discoverFromURL(browserCtx, searchURL, src.pattern, portalName, 0)
		if err != nil {
			// Best-effort: skip boards that fail.
			return
		}

		all = append(all, results...)
	})

	if config.MaxResults > 0 && len(all) > config.MaxResults {
		all = all[:config.MaxResults]
	}

	return all, nil
}

// --- Shared helpers ---

// queryParamsWithRemote builds url.Values with query and optional remote flag.
func queryParamsWithRemote(config BoardConfig) url.Values {
	params := url.Values{
		"query": {config.queryString()},
	}
	if config.Remote {
		params.Set("remote", "true")
	}
	return params
}

// discoverFromURL navigates to a URL, extracts links matching pattern,
// tags them with the portal name, and optionally caps results.
func discoverFromURL(
	browserCtx context.Context,
	pageURL, linkPattern, portalName string,
	maxResults int,
) ([]ScanResult, error) {
	results, err := extractFromBoard(browserCtx, pageURL, linkPattern)
	if err != nil {
		return nil, err
	}

	results = lo.Map(results, func(r ScanResult, _ int) ScanResult {
		r.Portal = portalName
		return r
	})

	if maxResults > 0 && len(results) > maxResults {
		results = results[:maxResults]
	}

	return results, nil
}

// extractFromBoard navigates to a URL in a new tab and extracts links matching the pattern.
func extractFromBoard(browserCtx context.Context, pageURL, linkPattern string) ([]ScanResult, error) {
	tabCtx, tabCancel := chromedp.NewContext(browserCtx)
	defer tabCancel()

	var links []jobLink
	if err := chromedp.Run(tabCtx,
		chromedp.Navigate(pageURL),
		chromedp.WaitReady("body"),
		chromedp.Evaluate(extractLinksJS, &links),
	); err != nil {
		return nil, oops.Wrapf(err, "extracting from %s", pageURL)
	}

	results := lo.FilterMap(links, func(l jobLink, _ int) (ScanResult, bool) {
		u := strings.TrimSpace(l.URL)
		title := strings.TrimSpace(l.Title)
		if u == "" || !strings.Contains(strings.ToLower(u), linkPattern) {
			return ScanResult{}, false
		}
		return ScanResult{URL: u, Title: title}, true
	})

	return results, nil
}
