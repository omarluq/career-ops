package db

import (
	"context"
	"time"

	"github.com/omarluq/career-ops/internal/model"
)

// Repository abstracts data access for career-ops pipeline data.
type Repository interface {
	// ListApplications returns all tracked applications ordered by number.
	ListApplications(ctx context.Context) ([]model.CareerApplication, error)

	// GetApplication returns a single application by its sequential number.
	GetApplication(ctx context.Context, number int) (*model.CareerApplication, error)

	// UpsertApplication creates or updates an application keyed by number.
	// If an application with the same company_key and role already exists, it
	// updates that row instead of inserting a duplicate.
	UpsertApplication(ctx context.Context, app *model.CareerApplication) error

	// DeleteApplication removes the application with the given number.
	DeleteApplication(ctx context.Context, number int) error

	// FindDuplicates returns clusters of applications that appear to be
	// duplicates based on normalized company name and fuzzy role matching.
	FindDuplicates(ctx context.Context) ([]DuplicateCluster, error)

	// SearchApplications performs a free-text search across company, role,
	// notes, and archetype fields.
	SearchApplications(ctx context.Context, query string) ([]model.CareerApplication, error)

	// AddToPipeline inserts a new URL into the processing pipeline.
	AddToPipeline(ctx context.Context, url, source string) error

	// ListPipeline returns pipeline entries, optionally filtered by status.
	// Pass an empty string to list all entries.
	ListPipeline(ctx context.Context, status string) ([]RepoPipelineEntry, error)

	// UpdatePipelineStatus transitions a pipeline entry to a new status.
	UpdatePipelineStatus(ctx context.Context, url, status string) error

	// RecordScan saves a scan-history entry for deduplication.
	RecordScan(ctx context.Context, url, portal, title, company string) error

	// HasBeenScanned checks whether a URL+portal combination was already scanned.
	HasBeenScanned(ctx context.Context, url, portal string) (bool, error)

	// ListScanHistory returns all scan-history records ordered by scan time.
	ListScanHistory(ctx context.Context) ([]RepoScanRecord, error)

	// ComputeMetrics calculates aggregate pipeline statistics.
	ComputeMetrics(ctx context.Context) (model.PipelineMetrics, error)

	// --- Profile ---

	// GetProfile returns the user profile, creating a default if none exists.
	GetProfile(ctx context.Context) (*model.UserProfile, error)

	// SaveProfile upserts the user profile.
	SaveProfile(ctx context.Context, p *model.UserProfile) error

	// UpdateProfileField updates a single field on the profile by name.
	UpdateProfileField(ctx context.Context, field, value string) error

	// --- Enrichments ---

	// RecordEnrichment stores a profile enrichment event.
	RecordEnrichment(ctx context.Context, e *model.ProfileEnrichment) error

	// ListPendingEnrichments returns enrichments not yet applied.
	ListPendingEnrichments(ctx context.Context) ([]model.ProfileEnrichment, error)

	// ApplyEnrichment marks an enrichment as applied.
	ApplyEnrichment(ctx context.Context, id int, appliedFields string) error

	// Close releases any resources held by the repository.
	Close() error
}

// DuplicateCluster groups applications that appear to be duplicates.
type DuplicateCluster struct {
	CompanyKey   string
	Applications []model.CareerApplication
}

// RepoPipelineEntry represents a URL queued for evaluation.
type RepoPipelineEntry struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	URL       string
	Source    string
	Status    string
}

// RepoScanRecord represents a single scan-history entry.
type RepoScanRecord struct {
	ScannedAt time.Time
	URL       string
	Portal    string
	Title     string
	Company   string
}
