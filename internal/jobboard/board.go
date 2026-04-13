// Package jobboard provides a unified interface for job board integrations,
// combining search (discover jobs) and apply (submit applications) into a
// single Board abstraction. It replaces the former scanner.BoardAdapter and
// applicator.PortalHandler split.
package jobboard

import "context"

// Board is the unified interface for job board integrations.
// Every board implements Search (discover jobs) and optionally Apply (submit applications).
type Board interface {
	// Meta returns static metadata about this board.
	Meta() BoardMeta
	// Search discovers job listings matching the given query.
	Search(ctx context.Context, query SearchQuery) ([]SearchResult, error)
	// Apply submits an application. Returns ErrApplyNotSupported if the board is search-only.
	Apply(ctx context.Context, app Application) (ApplyResult, error)
	// HealthCheck verifies the board is reachable. Used for status dashboards.
	HealthCheck(ctx context.Context) error
}
