package model

// UserProfile represents the user's career profile.
type UserProfile struct {
	Preferences      map[string]string
	GitHubURL        string
	PortfolioURL     string
	VisaStatus       string
	City             string
	Country          string
	FullName         string
	CompLocationFlex string
	Phone            string
	Location         string
	Timezone         string
	CompMinimum      string
	Email            string
	CompTarget       string
	TwitterURL       string
	Headline         string
	ExitStory        string
	LinkedInURL      string
	CompCurrency     string
	ProofPoints      []ProofPoint
	Superpowers      []string
	DealBreakers     []string
	Archetypes       []ArchetypeEntry
	TargetRoles      []string
	ID               int
}

// ProofPoint represents a portfolio item or accomplishment.
type ProofPoint struct {
	Name       string `json:"name"`
	URL        string `json:"url"`
	HeroMetric string `json:"hero_metric"`
}

// ArchetypeEntry represents a target role archetype.
type ArchetypeEntry struct {
	Name  string `json:"name"`
	Level string `json:"level"`
	Fit   string `json:"fit"` // primary, secondary, adjacent
}

// ProfileEnrichment represents a profile enrichment event.
type ProfileEnrichment struct {
	SourceType    string
	SourceURL     string
	SourceTitle   string
	ExtractedData string // JSON
	AppliedFields string // JSON array
	Confidence    string // low, medium, high
	AppliedAt     string
	CreatedAt     string
	ID            int
	Applied       bool
}
