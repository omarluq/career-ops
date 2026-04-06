package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/samber/oops"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var syncCheckPath string

var syncCheckCmd = &cobra.Command{
	Use:   "sync-check",
	Short: "Validate setup consistency across config files",
	Long: "Checks that required files exist, YAML configs parse correctly, " +
		"profile has required fields, and flags possible hardcoded metrics.",
	RunE: runSyncCheck,
}

func init() {
	syncCheckCmd.Flags().StringVar(
		&syncCheckPath, "path", ".", "path to the career-ops project root",
	)
}

// requiredFile describes a file that must exist for the setup to be valid.
type requiredFile struct {
	rel  string // path relative to project root
	desc string // human-readable description
}

func runSyncCheck(_ *cobra.Command, _ []string) error {
	root := syncCheckPath

	var warnings []string
	var errors []string

	checkRequiredFiles(root, &errors)
	checkCVLength(root, &warnings)
	validateYAML(root, &errors)
	checkProfile(root, &warnings)
	checkHardcodedMetrics(root, &warnings)
	checkDigestFreshness(root, &warnings)

	// ── Output ─────────────────────────────────────────────────────────
	fmt.Println()
	fmt.Println("=== career-ops sync check ===")
	fmt.Println()

	if len(errors) == 0 && len(warnings) == 0 {
		fmt.Println("All checks passed.")
	} else {
		if len(errors) > 0 {
			fmt.Printf("ERRORS (%d):\n", len(errors))
			for _, e := range errors {
				fmt.Printf("  ERROR: %s\n", e)
			}
		}
		if len(warnings) > 0 {
			fmt.Printf("\nWARNINGS (%d):\n", len(warnings))
			for _, w := range warnings {
				fmt.Printf("  WARN: %s\n", w)
			}
		}
	}

	fmt.Println()

	if len(errors) > 0 {
		return oops.Errorf("sync-check found %d error(s)", len(errors))
	}
	return nil
}

// checkRequiredFiles verifies that all required project files exist.
func checkRequiredFiles(root string, errors *[]string) {
	required := []requiredFile{
		{"cv.md", "CV in markdown format"},
		{"config/profile.yml",
			"Profile configuration (copy from config/profile.example.yml)"},
		{"portals.yml",
			"Portal/scanner configuration (copy from templates/portals.example.yml)"},
		{"data/applications.md", "Application tracker"},
		{"templates/states.yml", "Canonical status definitions"},
		{"templates/cv-template.html", "HTML template for CV/PDF generation"},
	}

	for _, f := range required {
		p := filepath.Join(root, f.rel)
		if _, err := os.Stat(p); os.IsNotExist(err) {
			*errors = append(*errors, fmt.Sprintf(
				"%s not found. %s", f.rel, f.desc,
			))
		}
	}
}

// checkCVLength warns if cv.md exists but is suspiciously short.
func checkCVLength(root string, warnings *[]string) {
	cvPath := filepath.Clean(filepath.Join(root, "cv.md"))
	data, err := os.ReadFile(cvPath)
	if err != nil {
		return
	}
	if len(strings.TrimSpace(string(data))) < 100 {
		*warnings = append(*warnings,
			"cv.md seems too short. Make sure it contains your full CV.")
	}
}

// validateYAML checks that all YAML config files parse without errors.
func validateYAML(root string, errors *[]string) {
	yamlFiles := []string{
		"config/profile.yml",
		"portals.yml",
		"templates/states.yml",
	}
	for _, rel := range yamlFiles {
		p := filepath.Clean(filepath.Join(root, rel))
		data, err := os.ReadFile(p)
		if err != nil {
			continue // already reported as missing above
		}
		var doc interface{}
		if unmarshalErr := yaml.Unmarshal(data, &doc); unmarshalErr != nil {
			*errors = append(*errors, fmt.Sprintf(
				"%s is not valid YAML: %v", rel, unmarshalErr,
			))
		}
	}
}

// checkProfile verifies that profile.yml has required fields and no example data.
func checkProfile(root string, warnings *[]string) {
	profilePath := filepath.Clean(filepath.Join(root, "config", "profile.yml"))
	data, err := os.ReadFile(profilePath)
	if err != nil {
		return
	}
	content := string(data)
	requiredFields := []string{"full_name", "email", "location"}
	for _, field := range requiredFields {
		if !strings.Contains(content, field) {
			*warnings = append(*warnings, fmt.Sprintf(
				"config/profile.yml is missing field: %s", field,
			))
		}
	}
	if strings.Contains(content, `"Jane Smith"`) ||
		strings.Contains(content, `Jane Smith`) {
		*warnings = append(*warnings,
			"config/profile.yml may still have example data (found 'Jane Smith').")
	}
}

// checkHardcodedMetrics scans prompt files for possible hardcoded numeric metrics.
func checkHardcodedMetrics(root string, warnings *[]string) {
	metricPattern := regexp.MustCompile(
		`(?i)\b\d{2,4}\+?\s*(hours?|%|evals?|layers?|tests?|fields?|bases?)\b`,
	)
	promptFiles := []struct {
		rel  string
		name string
	}{
		{"modes/_shared.md", "_shared.md"},
		{"batch/batch-prompt.md", "batch-prompt.md"},
	}

	for _, pf := range promptFiles {
		p := filepath.Clean(filepath.Join(root, pf.rel))
		fh, err := os.Open(p)
		if err != nil {
			continue
		}
		scanPromptFile(fh, pf.name, metricPattern, warnings)
		if closeErr := fh.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "warning: closing %s: %v\n", pf.name, closeErr)
		}
	}
}

// scanPromptFile scans a single file for hardcoded metric patterns.
func scanPromptFile(
	fh *os.File, name string, pattern *regexp.Regexp, warnings *[]string,
) {
	scanner := bufio.NewScanner(fh)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if strings.Contains(line, "NEVER hardcode") ||
			strings.Contains(line, "NUNCA hardcode") ||
			strings.HasPrefix(line, "#") ||
			strings.HasPrefix(line, "<!--") {
			continue
		}
		if matches := pattern.FindString(line); matches != "" {
			*warnings = append(*warnings, fmt.Sprintf(
				"%s:%d — Possible hardcoded metric: %q. "+
					"Should this be read from cv.md/article-digest.md?",
				name, lineNum, matches,
			))
		}
	}
}

// checkDigestFreshness warns if article-digest.md is older than 30 days.
func checkDigestFreshness(root string, warnings *[]string) {
	digestPath := filepath.Join(root, "article-digest.md")
	info, err := os.Stat(digestPath)
	if err != nil {
		return
	}
	daysSince := time.Since(info.ModTime()).Hours() / 24
	if daysSince > 30 {
		*warnings = append(*warnings, fmt.Sprintf(
			"article-digest.md is %d days old. "+
				"Consider updating if your projects have new metrics.",
			int(math.Round(daysSince)),
		))
	}
}
