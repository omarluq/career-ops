package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/samber/oops"
	"github.com/spf13/cobra"
)

var pdfFormat string

var pdfCmd = &cobra.Command{
	Use:   "pdf <input.html> <output.pdf>",
	Short: "Generate PDF from HTML template",
	Long:  "Renders an HTML CV template to PDF using headless Chrome (chromedp).",
	Args:  cobra.ExactArgs(2),
	RunE:  runPDF,
}

func init() {
	pdfCmd.Flags().StringVar(&pdfFormat, "format", "a4", "paper format: letter or a4")
}

// paperSize returns (width, height) in inches for the given format name.
func paperSize(format string) (w, h float64, err error) {
	switch strings.ToLower(format) {
	case "a4":
		return 8.27, 11.69, nil
	case "letter":
		return 8.5, 11.0, nil
	default:
		return 0, 0, oops.Errorf("invalid format %q, use: a4, letter", format)
	}
}

// rewriteFontURLs rewrites relative font paths in the HTML to absolute file:// URLs.
func rewriteFontURLs(html, inputDir string) string {
	// Resolve fonts/ relative to the input file's directory.
	fontsDir := filepath.Join(inputDir, "fonts")
	if _, statErr := os.Stat(fontsDir); statErr != nil {
		// Try CWD-relative fonts/ as well.
		if cwd, cwdErr := os.Getwd(); cwdErr == nil {
			candidate := filepath.Join(cwd, "fonts")
			if _, s2 := os.Stat(candidate); s2 == nil {
				fontsDir = candidate
			}
		}
	}
	html = strings.ReplaceAll(html, "url('./fonts/", "url('file://"+fontsDir+"/")
	html = strings.ReplaceAll(html, `url("./fonts/`, `url("file://`+fontsDir+"/")
	return html
}

// prepareHTML reads the input HTML, rewrites font URLs, and writes it to a temp file.
// Returns the file:// URL for the temp file and a cleanup function.
func prepareHTML(inputPath string) (fileURL string, cleanup func(), err error) {
	htmlBytes, err := os.ReadFile(filepath.Clean(inputPath))
	if err != nil {
		return "", nil, oops.Wrapf(err, "reading input HTML")
	}

	html := rewriteFontURLs(string(htmlBytes), filepath.Dir(inputPath))

	tmpFile, err := os.CreateTemp("", "career-ops-pdf-*.html")
	if err != nil {
		return "", nil, oops.Wrapf(err, "creating temp file")
	}
	cleanup = func() {
		if removeErr := os.Remove(tmpFile.Name()); removeErr != nil {
			fmt.Fprintf(os.Stderr, "warning: removing temp file: %v\n", removeErr)
		}
	}

	if _, writeErr := tmpFile.WriteString(html); writeErr != nil {
		if closeErr := tmpFile.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "warning: closing temp file: %v\n", closeErr)
		}
		cleanup()
		return "", nil, oops.Wrapf(writeErr, "writing temp HTML")
	}
	if closeErr := tmpFile.Close(); closeErr != nil {
		cleanup()
		return "", nil, oops.Wrapf(closeErr, "closing temp file")
	}

	return "file://" + tmpFile.Name(), cleanup, nil
}

// renderPDF launches headless Chrome and renders the HTML at fileURL to PDF bytes.
func renderPDF(fileURL string, width, height float64) ([]byte, error) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	const margin = 0.4 // inches

	var pdfBuf []byte
	if err := chromedp.Run(ctx,
		chromedp.Navigate(fileURL),
		chromedp.WaitReady("body"),
		chromedp.ActionFunc(func(ctx context.Context) error {
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
				Do(ctx)
			return printErr
		}),
	); err != nil {
		return nil, oops.Wrapf(err, "generating PDF")
	}
	return pdfBuf, nil
}

func runPDF(_ *cobra.Command, args []string) error {
	inputPath, err := filepath.Abs(args[0])
	if err != nil {
		return oops.Wrapf(err, "resolving input path")
	}
	outputPath, err := filepath.Abs(args[1])
	if err != nil {
		return oops.Wrapf(err, "resolving output path")
	}

	width, height, err := paperSize(pdfFormat)
	if err != nil {
		return err
	}

	if _, statErr := os.Stat(inputPath); statErr != nil {
		return oops.Wrapf(statErr, "input file")
	}

	fmt.Fprintf(os.Stderr, "Input:  %s\n", inputPath)
	fmt.Fprintf(os.Stderr, "Output: %s\n", outputPath)
	fmt.Fprintf(os.Stderr, "Format: %s\n", strings.ToUpper(pdfFormat))

	fileURL, cleanup, err := prepareHTML(inputPath)
	if err != nil {
		return err
	}
	defer cleanup()

	pdfBuf, err := renderPDF(fileURL, width, height)
	if err != nil {
		return err
	}

	if err = os.WriteFile(outputPath, pdfBuf, 0o600); err != nil {
		return oops.Wrapf(err, "writing PDF")
	}

	pageCount := strings.Count(string(pdfBuf), "/Type /Page") -
		strings.Count(string(pdfBuf), "/Type /Pages")
	if pageCount < 1 {
		pageCount = 1
	}

	fmt.Fprintf(os.Stderr, "PDF generated: %s\n", outputPath)
	fmt.Fprintf(os.Stderr, "Pages: %d\n", pageCount)
	fmt.Fprintf(os.Stderr, "Size:  %.1f KB\n", float64(len(pdfBuf))/1024)

	return nil
}
