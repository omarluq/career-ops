// Package scanner provides concurrent portal scanning for job discovery.
package scanner

import (
	"os"
	"path/filepath"

	"github.com/samber/lo"
	"github.com/samber/oops"
	"gopkg.in/yaml.v3"
)

// PortalConfig defines a portal to scan.
type PortalConfig struct {
	Name    string
	BaseURL string
	Queries []string
}

// TitleFilter holds keyword filters for job title relevance.
type TitleFilter struct {
	Positive       []string `yaml:"positive"`
	Negative       []string `yaml:"negative"`
	SeniorityBoost []string `yaml:"seniority_boost"`
}

// searchQuery is a single search query entry from portals.yml.
type searchQuery struct {
	Name    string `yaml:"name"`
	Query   string `yaml:"query"`
	Enabled bool   `yaml:"enabled"`
}

// trackedCompany is a single tracked company from portals.yml.
type trackedCompany struct {
	Name       string `yaml:"name"`
	CareersURL string `yaml:"careers_url"`
	API        string `yaml:"api"`
	ScanMethod string `yaml:"scan_method"`
	ScanQuery  string `yaml:"scan_query"`
	Notes      string `yaml:"notes"`
	Enabled    bool   `yaml:"enabled"`
}

// portalsFile is the top-level YAML structure of portals.yml.
type portalsFile struct {
	TitleFilter      TitleFilter      `yaml:"title_filter"`
	SearchQueries    []searchQuery    `yaml:"search_queries"`
	TrackedCompanies []trackedCompany `yaml:"tracked_companies"`
}

// LoadPortals reads portal configuration from the given YAML path
// and converts it to a slice of PortalConfig for scanning.
func LoadPortals(path string) ([]PortalConfig, *TitleFilter, error) {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, nil, oops.Wrapf(err, "reading portals file %s", path)
	}

	var pf portalsFile
	if err = yaml.Unmarshal(data, &pf); err != nil {
		return nil, nil, oops.Wrapf(err, "parsing portals YAML")
	}

	var portals []PortalConfig

	// Convert search queries into portal configs grouped by base domain.
	enabledQueries := lo.Filter(pf.SearchQueries, func(q searchQuery, _ int) bool {
		return q.Enabled
	})
	if len(enabledQueries) > 0 {
		portals = append(portals, PortalConfig{
			Name:    "search_queries",
			BaseURL: "",
			Queries: lo.Map(enabledQueries, func(q searchQuery, _ int) string {
				return q.Query
			}),
		})
	}

	// Convert tracked companies into portal configs.
	enabledCompanies := lo.Filter(pf.TrackedCompanies, func(c trackedCompany, _ int) bool {
		return c.Enabled
	})
	companyPortals := lo.Map(enabledCompanies, func(c trackedCompany, _ int) PortalConfig {
		pc := PortalConfig{
			Name:    c.Name,
			BaseURL: lo.Ternary(c.API != "", c.API, c.CareersURL),
		}
		if c.ScanQuery != "" {
			pc.Queries = []string{c.ScanQuery}
		}
		return pc
	})
	portals = append(portals, companyPortals...)

	return portals, &pf.TitleFilter, nil
}
