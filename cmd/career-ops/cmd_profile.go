package main

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/samber/lo"
	"github.com/samber/oops"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/omarluq/career-ops/internal/closer"
	"github.com/omarluq/career-ops/internal/db"
	"github.com/omarluq/career-ops/internal/model"
)

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Manage your career profile",
}

var profileShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Display your career profile",
	RunE:  runProfileShow,
}

var profileSetCmd = &cobra.Command{
	Use:   "set <field> <value>",
	Short: "Update a single profile field",
	Args:  cobra.MinimumNArgs(2),
	RunE:  runProfileSet,
}

var profileEnrichmentsCmd = &cobra.Command{
	Use:   "enrichments",
	Short: "List pending profile enrichments",
	RunE:  runProfileEnrichments,
}

var profileApplyCmd = &cobra.Command{
	Use:   "apply <id>",
	Short: "Apply a pending enrichment to your profile",
	Args:  cobra.ExactArgs(1),
	RunE:  runProfileApply,
}

func init() {
	profileCmd.AddCommand(
		profileShowCmd,
		profileSetCmd,
		profileEnrichmentsCmd,
		profileApplyCmd,
	)
}

// fieldAliases maps short CLI names to the canonical field names accepted
// by db.UpdateProfileField.
var fieldAliases = map[string]string{
	"name":          "full_name",
	"email":         "email",
	"phone":         "phone",
	"location":      "location",
	"timezone":      "timezone",
	"linkedin":      "linkedin_url",
	"github":        "github_url",
	"portfolio":     "portfolio_url",
	"twitter":       "twitter_url",
	"headline":      "headline",
	"exit-story":    "exit_story",
	"comp-target":   "comp_target",
	"comp-currency": "comp_currency",
	"comp-minimum":  "comp_minimum",
	"comp-flex":     "comp_location_flex",
	"country":       "country",
	"city":          "city",
	"visa":          "visa_status",
}

// profileWriter wraps an io.Writer and accumulates the first write error,
// following the same pattern as mdWriter in cmd_export.go.
type profileWriter struct {
	w   io.Writer
	err error
}

func (pw *profileWriter) writef(format string, args ...any) {
	if pw.err != nil {
		return
	}
	_, pw.err = fmt.Fprintf(pw.w, format, args...)
}

func runProfileShow(cmd *cobra.Command, _ []string) (err error) {
	ctx := cmd.Context()
	g := closer.Guard{Err: &err}

	database, err := db.OpenAndMigrate(ctx, viper.GetString("db"))
	if err != nil {
		return err
	}
	defer g.Close(database)

	r := db.NewSQLite(database)
	defer g.Close(r)

	p, err := r.GetProfile(ctx)
	if err != nil {
		return oops.Wrapf(err, "fetching profile")
	}

	pw := &profileWriter{w: cmd.OutOrStdout()}
	printProfileSection(pw, "Identity", profileRows(p, identityFields()))
	printProfileSection(pw, "Contact", profileRows(p, contactFields()))
	printProfileSection(pw, "Compensation", profileRows(p, compFields()))
	printProfileSection(pw, "Location", profileRows(p, locationFields()))
	printProfileSliceSection(pw, "Target Roles", p.TargetRoles)
	printProfileSliceSection(pw, "Superpowers", p.Superpowers)
	printProfileSliceSection(pw, "Deal Breakers", p.DealBreakers)
	printProfileArchetypes(pw, p.Archetypes)
	printProfileProofPoints(pw, p.ProofPoints)
	printProfilePreferences(pw, p.Preferences)

	if pw.err != nil {
		return oops.Wrapf(pw.err, "writing profile output")
	}
	return nil
}

func runProfileSet(cmd *cobra.Command, args []string) (err error) {
	ctx := cmd.Context()
	g := closer.Guard{Err: &err}

	field := args[0]
	value := strings.Join(args[1:], " ")

	canonical, ok := fieldAliases[field]
	if !ok {
		canonical = field
	}

	database, err := db.OpenAndMigrate(ctx, viper.GetString("db"))
	if err != nil {
		return err
	}
	defer g.Close(database)

	r := db.NewSQLite(database)
	defer g.Close(r)

	if err := r.UpdateProfileField(ctx, canonical, value); err != nil {
		return oops.Wrapf(err, "updating profile field %q", canonical)
	}

	_, printErr := fmt.Fprintf(cmd.OutOrStdout(),
		"Updated %s = %s\n", canonical, value)
	if printErr != nil {
		return oops.Wrapf(printErr, "writing confirmation")
	}
	return nil
}

func runProfileEnrichments(cmd *cobra.Command, _ []string) (err error) {
	ctx := cmd.Context()
	g := closer.Guard{Err: &err}

	database, err := db.OpenAndMigrate(ctx, viper.GetString("db"))
	if err != nil {
		return err
	}
	defer g.Close(database)

	r := db.NewSQLite(database)
	defer g.Close(r)

	enrichments, err := r.ListPendingEnrichments(ctx)
	if err != nil {
		return oops.Wrapf(err, "listing pending enrichments")
	}

	w := cmd.OutOrStdout()

	if len(enrichments) == 0 {
		_, printErr := fmt.Fprintln(w, "No pending enrichments.")
		if printErr != nil {
			return oops.Wrapf(printErr, "writing output")
		}
		return nil
	}

	pw := &profileWriter{w: w}
	lo.ForEach(enrichments, func(e model.ProfileEnrichment, _ int) {
		preview := e.ExtractedData
		if len(preview) > 80 {
			preview = preview[:80] + "..."
		}
		pw.writef("[%d] %s | %s | confidence: %s\n     data: %s\n\n",
			e.ID, e.SourceType, e.SourceTitle, e.Confidence, preview)
	})
	if pw.err != nil {
		return oops.Wrapf(pw.err, "writing enrichment list")
	}

	return nil
}

func runProfileApply(cmd *cobra.Command, args []string) (err error) {
	ctx := cmd.Context()
	g := closer.Guard{Err: &err}

	id, parseErr := strconv.Atoi(args[0])
	if parseErr != nil {
		return oops.Wrapf(parseErr, "parsing enrichment id %q", args[0])
	}

	database, err := db.OpenAndMigrate(ctx, viper.GetString("db"))
	if err != nil {
		return err
	}
	defer g.Close(database)

	r := db.NewSQLite(database)
	defer g.Close(r)

	if err := r.ApplyEnrichment(ctx, id, "all"); err != nil {
		return oops.Wrapf(err, "applying enrichment %d", id)
	}

	_, printErr := fmt.Fprintf(cmd.OutOrStdout(),
		"Applied enrichment %d\n", id)
	if printErr != nil {
		return oops.Wrapf(printErr, "writing confirmation")
	}
	return nil
}

// --- display helpers ---

type labelValue struct {
	Label string
	Value string
}

func printProfileSection(
	pw *profileWriter, title string, rows []labelValue,
) {
	filtered := lo.Filter(rows, func(r labelValue, _ int) bool {
		return r.Value != ""
	})
	if len(filtered) == 0 {
		return
	}
	pw.writef("\n=== %s ===\n", title)
	lo.ForEach(filtered, func(r labelValue, _ int) {
		pw.writef("  %-18s %s\n", r.Label+":", r.Value)
	})
}

func printProfileSliceSection(
	pw *profileWriter, title string, items []string,
) {
	if len(items) == 0 {
		return
	}
	pw.writef("\n=== %s ===\n", title)
	lo.ForEach(items, func(item string, _ int) {
		pw.writef("  - %s\n", item)
	})
}

func printProfileArchetypes(
	pw *profileWriter, archetypes []model.ArchetypeEntry,
) {
	if len(archetypes) == 0 {
		return
	}
	pw.writef("\n=== Archetypes ===\n")
	lo.ForEach(archetypes, func(a model.ArchetypeEntry, _ int) {
		pw.writef("  - %s (%s, %s)\n", a.Name, a.Level, a.Fit)
	})
}

func printProfileProofPoints(
	pw *profileWriter, points []model.ProofPoint,
) {
	if len(points) == 0 {
		return
	}
	pw.writef("\n=== Proof Points ===\n")
	lo.ForEach(points, func(pp model.ProofPoint, _ int) {
		line := fmt.Sprintf("  - %s", pp.Name)
		if pp.HeroMetric != "" {
			line += fmt.Sprintf(" [%s]", pp.HeroMetric)
		}
		if pp.URL != "" {
			line += fmt.Sprintf(" (%s)", pp.URL)
		}
		pw.writef("%s\n", line)
	})
}

func printProfilePreferences(
	pw *profileWriter, prefs map[string]string,
) {
	if len(prefs) == 0 {
		return
	}
	pw.writef("\n=== Preferences ===\n")
	lo.ForEach(lo.Keys(prefs), func(k string, _ int) {
		pw.writef("  %-18s %s\n", k+":", prefs[k])
	})
}

// --- field group definitions ---

type fieldDef struct {
	Getter func(*model.UserProfile) string
	Label  string
}

func identityFields() []fieldDef {
	return []fieldDef{
		{Label: "Name", Getter: func(p *model.UserProfile) string { return p.FullName }},
		{Label: "Headline", Getter: func(p *model.UserProfile) string { return p.Headline }},
		{Label: "Exit Story", Getter: func(p *model.UserProfile) string { return p.ExitStory }},
	}
}

func contactFields() []fieldDef {
	return []fieldDef{
		{Label: "Email", Getter: func(p *model.UserProfile) string { return p.Email }},
		{Label: "Phone", Getter: func(p *model.UserProfile) string { return p.Phone }},
		{Label: "LinkedIn", Getter: func(p *model.UserProfile) string { return p.LinkedInURL }},
		{Label: "GitHub", Getter: func(p *model.UserProfile) string { return p.GitHubURL }},
		{Label: "Portfolio", Getter: func(p *model.UserProfile) string { return p.PortfolioURL }},
		{Label: "Twitter", Getter: func(p *model.UserProfile) string { return p.TwitterURL }},
	}
}

func compFields() []fieldDef {
	return []fieldDef{
		{Label: "Target", Getter: func(p *model.UserProfile) string { return p.CompTarget }},
		{Label: "Currency", Getter: func(p *model.UserProfile) string { return p.CompCurrency }},
		{Label: "Minimum", Getter: func(p *model.UserProfile) string { return p.CompMinimum }},
		{Label: "Location Flex", Getter: func(p *model.UserProfile) string { return p.CompLocationFlex }},
	}
}

func locationFields() []fieldDef {
	return []fieldDef{
		{Label: "Location", Getter: func(p *model.UserProfile) string { return p.Location }},
		{Label: "Country", Getter: func(p *model.UserProfile) string { return p.Country }},
		{Label: "City", Getter: func(p *model.UserProfile) string { return p.City }},
		{Label: "Timezone", Getter: func(p *model.UserProfile) string { return p.Timezone }},
		{Label: "Visa Status", Getter: func(p *model.UserProfile) string { return p.VisaStatus }},
	}
}

func profileRows(p *model.UserProfile, defs []fieldDef) []labelValue {
	return lo.Map(defs, func(d fieldDef, _ int) labelValue {
		return labelValue{Label: d.Label, Value: d.Getter(p)}
	})
}
