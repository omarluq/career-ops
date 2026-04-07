package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/samber/lo"
	"github.com/samber/mo"
	"github.com/samber/oops"
	"github.com/spf13/cobra"

	"github.com/omarluq/career-ops/internal/worker"
)

var (
	pdfBatchConcurrency int
	pdfBatchFormat      string
)

var pdfBatchCmd = &cobra.Command{
	Use:   "pdf-batch <glob-or-dir>",
	Short: "Generate PDFs from multiple HTML files concurrently",
	Long: "Processes all matching HTML files in parallel using a shared Chrome allocator.\n" +
		"Output PDFs are written alongside each input with a .pdf extension.",
	Args: cobra.ExactArgs(1),
	RunE: runPDFBatch,
}

func init() {
	pdfBatchCmd.Flags().IntVar(&pdfBatchConcurrency, "concurrency", 3, "number of concurrent Chrome tabs")
	pdfBatchCmd.Flags().StringVar(&pdfBatchFormat, "format", "a4", "paper format: letter or a4")
}

// pdfBatchItem represents a single HTML-to-PDF conversion job.
type pdfBatchItem struct {
	InputPath  string
	OutputPath string
}

// pdfBatchResult holds the outcome of one conversion.
type pdfBatchResult struct {
	InputPath  string
	OutputPath string
	Pages      int
	SizeKB     float64
}

func runPDFBatch(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	width, height, err := paperSize(pdfBatchFormat)
	if err != nil {
		return err
	}

	// Resolve input files.
	pattern := args[0]
	files, err := resolveHTMLFiles(pattern)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return oops.Errorf("no HTML files matched pattern %q", pattern)
	}

	stderrf("Found %d HTML file(s), concurrency=%d\n", len(files), pdfBatchConcurrency)

	// Build batch items.
	items := lo.Map(files, func(f string, _ int) pdfBatchItem {
		return pdfBatchItem{
			InputPath:  f,
			OutputPath: strings.TrimSuffix(f, filepath.Ext(f)) + ".pdf",
		}
	})

	// Create shared Chrome allocator.
	allocCtx, allocCancel := chromedp.NewExecAllocator(ctx, chromedp.DefaultExecAllocatorOptions[:]...)
	defer allocCancel()

	browserCtx, browserCancel := chromedp.NewContext(allocCtx)
	defer browserCancel()

	if err = chromedp.Run(browserCtx); err != nil {
		return oops.Wrapf(err, "starting browser for batch PDF")
	}

	// Process all files concurrently via worker pool.
	results, batchErr := worker.RunBatch(ctx, worker.BatchConfig[pdfBatchItem, pdfBatchResult]{
		Items:       items,
		Concurrency: pdfBatchConcurrency,
		Processor: func(_ context.Context, item pdfBatchItem) (pdfBatchResult, error) {
			return processSinglePDF(browserCtx, item, width, height)
		},
		OnProgress: func(completed, total int) {
			stderrf("  [%d/%d] complete\n", completed, total)
		},
	})

	// Report results.
	var succeeded, failed int
	lo.ForEach(results, func(r mo.Result[pdfBatchResult], i int) {
		val, getErr := r.Get()
		if getErr != nil {
			failed++
			stderrf("FAIL: %s -- %v\n", items[i].InputPath, getErr)
			return
		}
		succeeded++
		stderrf("OK:   %s -> %s (%d pages, %.1f KB)\n",
			val.InputPath, val.OutputPath, val.Pages, val.SizeKB)
	})

	stderrf("\nBatch complete: %d succeeded, %d failed\n", succeeded, failed)

	if batchErr != nil {
		return oops.Wrapf(batchErr, "batch PDF generation had errors")
	}
	return nil
}

// processSinglePDF converts one HTML file to PDF using a tab from the shared browser.
func processSinglePDF(
	browserCtx context.Context,
	item pdfBatchItem,
	width, height float64,
) (pdfBatchResult, error) {
	fileURL, cleanup, err := prepareHTML(item.InputPath)
	if err != nil {
		return pdfBatchResult{}, oops.Wrapf(err, "preparing %s", item.InputPath)
	}
	defer cleanup()

	// Create a new tab within the shared browser.
	tabCtx, tabCancel := chromedp.NewContext(browserCtx)
	defer tabCancel()

	pdfBuf, err := renderPDFWithCtx(tabCtx, fileURL, width, height)
	if err != nil {
		return pdfBatchResult{}, oops.Wrapf(err, "rendering %s", item.InputPath)
	}

	if err = os.WriteFile(item.OutputPath, pdfBuf, 0o600); err != nil {
		return pdfBatchResult{}, oops.Wrapf(err, "writing %s", item.OutputPath)
	}

	pageCount := strings.Count(string(pdfBuf), "/Type /Page") -
		strings.Count(string(pdfBuf), "/Type /Pages")
	if pageCount < 1 {
		pageCount = 1
	}

	return pdfBatchResult{
		InputPath:  item.InputPath,
		OutputPath: item.OutputPath,
		Pages:      pageCount,
		SizeKB:     float64(len(pdfBuf)) / 1024,
	}, nil
}

// renderPDFWithCtx renders a PDF using the provided chromedp context (tab).
func renderPDFWithCtx(ctx context.Context, fileURL string, width, height float64) ([]byte, error) {
	const margin = 0.4

	var pdfBuf []byte
	if err := chromedp.Run(ctx,
		chromedp.Navigate(fileURL),
		chromedp.WaitReady("body"),
		chromedp.ActionFunc(func(actCtx context.Context) error {
			var printErr error
			pdfBuf, _, printErr = page.PrintToPDF().
				WithPaperWidth(width).
				WithPaperHeight(height).
				WithMarginTop(margin).
				WithMarginRight(margin).
				WithMarginBottom(margin).
				WithMarginLeft(margin).
				WithPrintBackground(true).
				WithPreferCSSPageSize(false).
				Do(actCtx)
			return printErr
		}),
	); err != nil {
		return nil, oops.Wrapf(err, "generating PDF from %s", fileURL)
	}
	return pdfBuf, nil
}

// resolveHTMLFiles expands a glob pattern or directory into a list of HTML file paths.
func resolveHTMLFiles(pattern string) ([]string, error) {
	info, err := os.Stat(pattern)
	if err == nil && info.IsDir() {
		pattern = filepath.Join(pattern, "*.html")
	}

	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, oops.Wrapf(err, "expanding glob pattern %q", pattern)
	}

	htmlFiles := lo.FilterMap(matches, func(m string, _ int) (string, bool) {
		if !strings.HasSuffix(strings.ToLower(m), ".html") {
			return "", false
		}
		abs, absErr := filepath.Abs(m)
		if absErr != nil {
			return "", false
		}
		return abs, true
	})

	return htmlFiles, nil
}
