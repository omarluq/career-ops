package boards

import (
	"context"
	"net/url"
	"strings"
	"time"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// BaseBoard provides default implementations of Meta, HealthCheck, and Apply
// for boards that follow the standard pattern. Embed it in concrete board
// structs to eliminate boilerplate — only override Search (required) and
// Apply (if the board supports it).
type BaseBoard struct {
	BMeta jobboard.BoardMeta
}

// Meta returns the board metadata.
func (b *BaseBoard) Meta() jobboard.BoardMeta { return b.BMeta }

// HealthCheck verifies the board is reachable via HEAD request.
func (b *BaseBoard) HealthCheck(ctx context.Context) error {
	return healthCheck(ctx, b.BMeta.URL)
}

// Apply returns ErrApplyNotSupported. Override in boards that support applications.
func (b *BaseBoard) Apply(ctx context.Context, app jobboard.Application) (jobboard.ApplyResult, error) {
	return applyNotSupported(ctx, app)
}

// ScrapeBoard is a fully-configured Board for sites that use buildSearchURL +
// scrapeJobLinks. No custom struct or Search override needed — just provide
// the metadata, base search URL, and link-match pattern.
type ScrapeBoard struct {
	searchBase  string
	linkPattern string
	BaseBoard
	remoteAware bool
}

// NewScrapeBoard creates a Board that searches via URL scraping.
func NewScrapeBoard(
	meta jobboard.BoardMeta, searchBase, linkPattern string,
) *ScrapeBoard {
	return &ScrapeBoard{
		BaseBoard:   BaseBoard{BMeta: meta},
		searchBase:  searchBase,
		linkPattern: linkPattern,
	}
}

// NewRemoteScrapeBoard creates a ScrapeBoard that appends a remote filter
// query parameter when the caller sets q.Remote = true.
func NewRemoteScrapeBoard(
	meta jobboard.BoardMeta, searchBase, linkPattern string,
) *ScrapeBoard {
	return &ScrapeBoard{
		BaseBoard:   BaseBoard{BMeta: meta},
		searchBase:  searchBase,
		linkPattern: linkPattern,
		remoteAware: true,
	}
}

// Search builds a keyword URL and scrapes matching job links from the page.
func (b *ScrapeBoard) Search(
	ctx context.Context, q jobboard.SearchQuery,
) ([]jobboard.SearchResult, error) {
	searchURL := buildSearchURL(b.searchBase, q.Keywords)
	if b.remoteAware && q.Remote {
		searchURL += remoteQueryParam
	}

	return scrapeJobLinks(ctx, searchURL, b.linkPattern, b.BMeta.Slug, defaultMax(q.MaxResults))
}

// ApplyScrapeBoard wraps ScrapeBoard for boards that advertise CapApply but
// have not yet implemented real submission — Apply returns ApplySkipped.
type ApplyScrapeBoard struct{ *ScrapeBoard }

// NewApplyScrapeBoard creates a ScrapeBoard whose Apply returns ApplySkipped.
func NewApplyScrapeBoard(
	meta jobboard.BoardMeta, searchBase, linkPattern string,
) *ApplyScrapeBoard {
	return &ApplyScrapeBoard{ScrapeBoard: NewScrapeBoard(meta, searchBase, linkPattern)}
}

// Apply returns a skipped result (stub for future implementation).
func (b *ApplyScrapeBoard) Apply(_ context.Context, _ jobboard.Application) (jobboard.ApplyResult, error) {
	return jobboard.ApplyResult{Status: jobboard.ApplySkipped, SubmittedAt: time.Now()}, nil
}

// SearchParams configures how keyword, location, and remote values map to
// query-string parameters when building search URLs.
type SearchParams struct {
	KeywordParam  string // e.g. "q", "keywords", "what"
	LocationParam string // e.g. "l", "location", "where" (empty = skip)
	RemoteParam   string // empty = no remote filter
	RemoteValue   string // e.g. "true", "1", "only_remote"
}

// ParamScrapeBoard builds search URLs using url.Values (keyword + optional
// location/remote params) and then scrapes job links from the result page.
type ParamScrapeBoard struct {
	searchBase  string
	linkPattern string
	params      SearchParams
	BaseBoard
}

// NewParamScrapeBoard creates a Board that searches via parameterised URL scraping.
func NewParamScrapeBoard(
	meta jobboard.BoardMeta, searchBase, linkPattern string, params SearchParams,
) *ParamScrapeBoard {
	return &ParamScrapeBoard{
		BaseBoard:   BaseBoard{BMeta: meta},
		searchBase:  searchBase,
		linkPattern: linkPattern,
		params:      params,
	}
}

// Search builds a parameterised search URL and scrapes matching job links.
func (b *ParamScrapeBoard) Search(
	ctx context.Context, q jobboard.SearchQuery,
) ([]jobboard.SearchResult, error) {
	params := url.Values{b.params.KeywordParam: {strings.Join(q.Keywords, " ")}}
	if q.Location != "" && b.params.LocationParam != "" {
		params.Set(b.params.LocationParam, q.Location)
	}
	if q.Remote && b.params.RemoteParam != "" {
		params.Set(b.params.RemoteParam, b.params.RemoteValue)
	}

	searchURL := b.searchBase + "?" + params.Encode()

	return scrapeJobLinks(ctx, searchURL, b.linkPattern, b.BMeta.Slug, defaultMax(q.MaxResults))
}
