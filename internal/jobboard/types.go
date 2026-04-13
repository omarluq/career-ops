package jobboard

import "time"

// BoardCategory classifies a job board by its primary function.
type BoardCategory string

const (
	// CategoryATS represents applicant tracking systems (Greenhouse, Lever, Ashby, Workable).
	CategoryATS BoardCategory = "ats"
	// CategoryAggregator represents large aggregators (LinkedIn, Indeed, Glassdoor).
	CategoryAggregator BoardCategory = "aggregator"
	// CategoryNiche represents niche/vertical boards (AI-Jobs, RemoteOK, WeWorkRemotely).
	CategoryNiche BoardCategory = "niche"
	// CategoryStartup represents startup-focused boards (YC, Wellfound, Underdog).
	CategoryStartup BoardCategory = "startup"
	// CategoryFreelance represents freelance platforms (Toptal, Upwork).
	CategoryFreelance BoardCategory = "freelance"
)

// Capability describes a feature supported by a board.
type Capability string

const (
	// CapSearch indicates the board supports job discovery.
	CapSearch Capability = "search"
	// CapApply indicates the board supports application submission.
	CapApply Capability = "apply"
	// CapAPI indicates the board exposes a structured API.
	CapAPI Capability = "api"
	// CapScrape indicates the board requires browser scraping.
	CapScrape Capability = "scrape"
)

// AuthType describes the authentication mechanism required by a board.
type AuthType string

const (
	// AuthNone means no authentication is required.
	AuthNone AuthType = "none"
	// AuthAPIKey means the board requires an API key.
	AuthAPIKey AuthType = "api_key"
	// AuthOAuth means the board uses OAuth authentication.
	AuthOAuth AuthType = "oauth"
	// AuthSession means the board requires a session cookie.
	AuthSession AuthType = "session"
)

// ApplyStatus describes the outcome of an application submission attempt.
type ApplyStatus string

const (
	// ApplySubmitted means the application was successfully submitted.
	ApplySubmitted ApplyStatus = "submitted"
	// ApplyFailed means the submission encountered an error.
	ApplyFailed ApplyStatus = "failed"
	// ApplySkipped means the submission was intentionally skipped.
	ApplySkipped ApplyStatus = "skipped"
	// ApplyPending means the submission is queued but not yet confirmed.
	ApplyPending ApplyStatus = "pending"
)

// RateConfig holds per-board rate limiting parameters.
type RateConfig struct {
	RequestsPerMinute int
	BurstSize         int
	CooldownOnError   time.Duration
}

// BoardMeta holds static information about a job board.
type BoardMeta struct {
	URL          string
	Name         string
	Slug         string
	Category     BoardCategory
	AuthType     AuthType
	Capabilities []Capability
	RateLimit    RateConfig
}

// SearchQuery describes a job search request.
type SearchQuery struct {
	Location   string
	PageToken  string
	Keywords   []string
	MaxResults int
	Remote     bool
}

// SearchResult represents a single job listing discovered by a board.
type SearchResult struct {
	PostedAt time.Time
	URL      string
	Title    string
	Company  string
	Location string
	Board    string
	Remote   bool
}

// Application holds all data needed to submit a job application.
type Application struct {
	FormData    map[string]string
	JDURL       string
	CVPDF       string
	CoverLetter string
	Company     string
	Role        string
}

// ApplyResult describes the outcome of an application submission.
type ApplyResult struct {
	SubmittedAt  time.Time
	ConfirmURL   string
	ErrorMessage string
	Status       ApplyStatus
	Duration     time.Duration
}
