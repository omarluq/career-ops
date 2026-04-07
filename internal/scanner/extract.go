package scanner

import (
	"context"
	"strings"

	"github.com/chromedp/chromedp"
	"github.com/samber/lo"
	"github.com/samber/oops"
)

// jobLink holds a raw link extracted from a page before enrichment.
type jobLink struct {
	URL   string
	Title string
}

// ExtractListings extracts job listing URLs and titles from a portal page.
// It navigates to the URL with chromedp and scrapes links matching common
// job listing patterns (greenhouse, ashby, lever, workable, etc.).
func ExtractListings(ctx context.Context, pageURL string) ([]ScanResult, error) {
	tabCtx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	var links []jobLink
	if err := chromedp.Run(tabCtx,
		chromedp.Navigate(pageURL),
		chromedp.WaitReady("body"),
		chromedp.Evaluate(extractLinksJS, &links),
	); err != nil {
		return nil, oops.Wrapf(err, "extracting listings from %s", pageURL)
	}

	results := lo.FilterMap(links, func(l jobLink, _ int) (ScanResult, bool) {
		url := strings.TrimSpace(l.URL)
		title := strings.TrimSpace(l.Title)
		if url == "" || !isJobURL(url) {
			return ScanResult{}, false
		}
		return ScanResult{URL: url, Title: title}, true
	})

	return results, nil
}

// isJobURL checks whether a URL looks like a job listing link.
func isJobURL(u string) bool {
	patterns := []string{
		"greenhouse.io",
		"ashbyhq.com",
		"lever.co",
		"workable.com",
		"jobs.",
		"careers",
		"/job/",
		"/jobs/",
		"/position/",
		"/opening/",
	}
	lower := strings.ToLower(u)
	return lo.SomeBy(patterns, func(p string) bool {
		return strings.Contains(lower, p)
	})
}

// extractLinksJS is JavaScript that extracts all anchor links from the page.
const extractLinksJS = `
(function() {
	var anchors = document.querySelectorAll('a[href]');
	var results = [];
	for (var i = 0; i < anchors.length; i++) {
		var a = anchors[i];
		var href = a.href || '';
		var text = (a.textContent || '').trim();
		if (href && text) {
			results.push({URL: href, Title: text});
		}
	}
	return results;
})()
`
