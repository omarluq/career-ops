package applicator

import (
	"context"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/samber/lo"
	"github.com/samber/oops"
)

// PortalHandler defines how to submit an application to a specific ATS.
type PortalHandler interface {
	// Portal returns the portal type this handler supports.
	Portal() PortalType
	// DetectPortal reports whether the given URL belongs to this ATS.
	DetectPortal(url string) bool
	// Submit executes the application submission in the given browser context.
	Submit(ctx context.Context, browserCtx context.Context, job SubmissionJob) (SubmissionResult, error)
}

// portalDetector pairs a domain fragment with its portal type for URL detection.
type portalDetector struct {
	fragment string
	portal   PortalType
}

// detectors is the ordered list of URL fragment matchers.
var detectors = []portalDetector{
	{fragment: "greenhouse.io", portal: PortalGreenhouse},
	{fragment: "lever.co", portal: PortalLever},
	{fragment: "workable.com", portal: PortalWorkable},
	{fragment: "linkedin.com", portal: PortalLinkedIn},
	{fragment: "ashbyhq.com", portal: PortalAshby},
}

// DetectPortal determines which ATS portal a URL belongs to.
// Falls back to PortalGeneric when no known pattern matches.
func DetectPortal(url string) PortalType {
	lower := strings.ToLower(url)

	match, found := lo.Find(detectors, func(d portalDetector) bool {
		return strings.Contains(lower, d.fragment)
	})
	if found {
		return match.portal
	}

	return PortalGeneric
}

// AllHandlers returns one handler instance per supported portal type.
func AllHandlers() []PortalHandler {
	return []PortalHandler{
		&GreenhouseHandler{},
		&LeverHandler{},
		&WorkableHandler{},
		&LinkedInHandler{},
		&AshbyHandler{},
		&GenericHandler{},
	}
}

// HandlerFor returns the first handler whose DetectPortal matches the URL.
func HandlerFor(url string) PortalHandler {
	handlers := AllHandlers()
	match, found := lo.Find(handlers, func(h PortalHandler) bool {
		return h.DetectPortal(url)
	})
	if found {
		return match
	}

	return &GenericHandler{}
}

// --- Stub handlers ---
// Each handler navigates to the URL, logs the portal type detected,
// and returns StatusSkipped until real form-filling logic is added.

// stubSubmit is the shared submission logic for all stub handlers.
// It navigates to the job URL, captures the page title, and returns
// a skipped result indicating the handler is not yet implemented.
func stubSubmit(
	_ context.Context,
	browserCtx context.Context,
	job SubmissionJob,
	portal PortalType,
) (SubmissionResult, error) {
	start := time.Now()

	tabCtx, tabCancel := chromedp.NewContext(browserCtx)
	defer tabCancel()

	var pageTitle string
	err := chromedp.Run(tabCtx,
		chromedp.Navigate(job.JDURL),
		chromedp.Title(&pageTitle),
	)
	if err != nil {
		return SubmissionResult{
			Job:          job,
			Status:       StatusFailed,
			ErrorMessage: oops.Wrapf(err, "navigating to %s", job.JDURL).Error(),
			SubmittedAt:  time.Now(),
			Duration:     time.Since(start),
		}, oops.Wrapf(err, "stub submit for portal %s", portal)
	}

	return SubmissionResult{
		Job:          job,
		Status:       StatusSkipped,
		ErrorMessage: "portal handler not yet implemented",
		SubmittedAt:  time.Now(),
		Duration:     time.Since(start),
	}, nil
}

// GreenhouseHandler handles greenhouse.io application forms.
type GreenhouseHandler struct{}

// Portal implements PortalHandler.
func (h *GreenhouseHandler) Portal() PortalType { return PortalGreenhouse }

// DetectPortal implements PortalHandler.
func (h *GreenhouseHandler) DetectPortal(u string) bool {
	return strings.Contains(strings.ToLower(u), "greenhouse.io")
}

// Submit implements PortalHandler.
func (h *GreenhouseHandler) Submit(
	ctx context.Context, browserCtx context.Context, job SubmissionJob,
) (SubmissionResult, error) {
	return stubSubmit(ctx, browserCtx, job, PortalGreenhouse)
}

// LeverHandler handles lever.co application forms.
type LeverHandler struct{}

// Portal implements PortalHandler.
func (h *LeverHandler) Portal() PortalType { return PortalLever }

// DetectPortal implements PortalHandler.
func (h *LeverHandler) DetectPortal(u string) bool {
	return strings.Contains(strings.ToLower(u), "lever.co")
}

// Submit implements PortalHandler.
func (h *LeverHandler) Submit(
	ctx context.Context, browserCtx context.Context, job SubmissionJob,
) (SubmissionResult, error) {
	return stubSubmit(ctx, browserCtx, job, PortalLever)
}

// WorkableHandler handles workable.com application forms.
type WorkableHandler struct{}

// Portal implements PortalHandler.
func (h *WorkableHandler) Portal() PortalType { return PortalWorkable }

// DetectPortal implements PortalHandler.
func (h *WorkableHandler) DetectPortal(u string) bool {
	return strings.Contains(strings.ToLower(u), "workable.com")
}

// Submit implements PortalHandler.
func (h *WorkableHandler) Submit(
	ctx context.Context, browserCtx context.Context, job SubmissionJob,
) (SubmissionResult, error) {
	return stubSubmit(ctx, browserCtx, job, PortalWorkable)
}

// LinkedInHandler handles linkedin.com application forms.
type LinkedInHandler struct{}

// Portal implements PortalHandler.
func (h *LinkedInHandler) Portal() PortalType { return PortalLinkedIn }

// DetectPortal implements PortalHandler.
func (h *LinkedInHandler) DetectPortal(u string) bool {
	return strings.Contains(strings.ToLower(u), "linkedin.com")
}

// Submit implements PortalHandler.
func (h *LinkedInHandler) Submit(
	ctx context.Context, browserCtx context.Context, job SubmissionJob,
) (SubmissionResult, error) {
	return stubSubmit(ctx, browserCtx, job, PortalLinkedIn)
}

// AshbyHandler handles ashbyhq.com application forms.
type AshbyHandler struct{}

// Portal implements PortalHandler.
func (h *AshbyHandler) Portal() PortalType { return PortalAshby }

// DetectPortal implements PortalHandler.
func (h *AshbyHandler) DetectPortal(u string) bool {
	return strings.Contains(strings.ToLower(u), "ashbyhq.com")
}

// Submit implements PortalHandler.
func (h *AshbyHandler) Submit(
	ctx context.Context, browserCtx context.Context, job SubmissionJob,
) (SubmissionResult, error) {
	return stubSubmit(ctx, browserCtx, job, PortalAshby)
}

// GenericHandler is the fallback for unknown ATS portals.
type GenericHandler struct{}

// Portal implements PortalHandler.
func (h *GenericHandler) Portal() PortalType { return PortalGeneric }

// DetectPortal implements PortalHandler.
func (h *GenericHandler) DetectPortal(_ string) bool { return true }

// Submit implements PortalHandler.
func (h *GenericHandler) Submit(
	ctx context.Context, browserCtx context.Context, job SubmissionJob,
) (SubmissionResult, error) {
	return stubSubmit(ctx, browserCtx, job, PortalGeneric)
}
