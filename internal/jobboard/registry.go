package jobboard

import (
	"context"
	"sync"

	"github.com/samber/lo"
	"github.com/samber/oops"
	"golang.org/x/sync/errgroup"
)

// Registry holds all registered boards and provides lookup by slug,
// category, and capability. It is safe for concurrent use.
type Registry struct {
	boards map[string]Board
	mu     sync.RWMutex
}

// NewRegistry creates an empty board registry.
func NewRegistry() *Registry {
	return &Registry{
		boards: make(map[string]Board),
	}
}

// Register adds a board to the registry, keyed by its slug.
// If a board with the same slug already exists it is replaced.
func (r *Registry) Register(b Board) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.boards[b.Meta().Slug] = b
}

// Get returns the board with the given slug, or false if not found.
func (r *Registry) Get(slug string) (Board, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	b, ok := r.boards[slug]

	return b, ok
}

// All returns every registered board in no guaranteed order.
func (r *Registry) All() []Board {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return lo.Values(r.boards)
}

// ByCategory returns all boards matching the given category.
func (r *Registry) ByCategory(cat BoardCategory) []Board {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return lo.Filter(lo.Values(r.boards), func(b Board, _ int) bool {
		return b.Meta().Category == cat
	})
}

// WithCapability returns all boards that advertise the given capability.
func (r *Registry) WithCapability(capability Capability) []Board {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return lo.Filter(lo.Values(r.boards), func(b Board, _ int) bool {
		return lo.Contains(b.Meta().Capabilities, capability)
	})
}

// SearchAll fans out the query to every board with CapSearch in parallel,
// collects results, and deduplicates by URL (first occurrence wins).
func (r *Registry) SearchAll(ctx context.Context, query SearchQuery) ([]SearchResult, error) {
	searchable := r.WithCapability(CapSearch)
	if len(searchable) == 0 {
		return nil, nil
	}

	var (
		resultsMu sync.Mutex
		all       []SearchResult
	)

	g, gCtx := errgroup.WithContext(ctx)

	lo.ForEach(searchable, func(b Board, _ int) {
		g.Go(func() error {
			results, err := b.Search(gCtx, query)
			if err != nil {
				return oops.
					In("jobboard").
					Tags("search-all", b.Meta().Slug).
					Wrapf(err, "searching board %s", b.Meta().Slug)
			}

			resultsMu.Lock()
			all = append(all, results...)
			resultsMu.Unlock()

			return nil
		})
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	// Deduplicate by URL — first occurrence wins.
	deduped := lo.UniqBy(all, func(r SearchResult) string {
		return r.URL
	})

	return deduped, nil
}
