package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
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
func paperSize(format string) (float64, float64, error) {
	switch strings.ToLower(format) {
	case "a4":
		return 8.27, 11.69, nil
	case "letter":
		return 8.5, 11.0, nil
	default:
		return 0, 0, fmt.Errorf("invalid format %q, use: a4, letter", format)
	}
}

func runPDF(cmd *cobra.Command, args []string) error {
	inputPath, err := filepath.Abs(args[0])
	if err != nil {
		return fmt.Errorf("resolving input path: %w", err)
	}
	outputPath, err := filepath.Abs(args[1])
	if err != nil {
		return fmt.Errorf("resolving output path: %w", err)
	}

	width, height, err := paperSize(pdfFormat)
	if err != nil {
		return err
	}

	// Verify input file exists.
	if _, err := os.Stat(inputPath); err != nil {
		return fmt.Errorf("input file: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Input:  %s\n", inputPath)
	fmt.Fprintf(os.Stderr, "Output: %s\n", outputPath)
	fmt.Fprintf(os.Stderr, "Format: %s\n", strings.ToUpper(pdfFormat))

	// Read HTML and rewrite relative font paths to absolute file:// URLs,
	// mirroring the behaviour of generate-pdf.mjs.
	htmlBytes, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("reading input HTML: %w", err)
	}
	html := string(htmlBytes)

	// Resolve fonts/ relative to the project root (where the binary is
	// typically invoked). Fall back to the input file's directory.
	fontsDir := filepath.Join(filepath.Dir(inputPath), "fonts")
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

	// Write the modified HTML to a temp file so chromedp can navigate to it.
	tmpFile, err := os.CreateTemp("", "career-ops-pdf-*.html")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(html); err != nil {
		tmpFile.Close()
		return fmt.Errorf("writing temp HTML: %w", err)
	}
	tmpFile.Close()

	fileURL := "file://" + tmpFile.Name()

	// Launch headless Chrome via chromedp.
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	const margin = 0.4 // inches

	var pdfBuf []byte
	if err := chromedp.Run(ctx,
		chromedp.Navigate(fileURL),
		chromedp.WaitReady("body"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			pdfBuf, _, err = page.PrintToPDF().
				WithPaperWidth(width).
				WithPaperHeight(height).
				WithMarginTop(margin).
				WithMarginRight(margin).
				WithMarginBottom(margin).
				WithMarginLeft(margin).
				WithPrintBackground(true).
				WithPreferCSSPageSize(false).
				Do(ctx)
			return err
		}),
	); err != nil {
		return fmt.Errorf("generating PDF: %w", err)
	}

	if err := os.WriteFile(outputPath, pdfBuf, 0o644); err != nil {
		return fmt.Errorf("writing PDF: %w", err)
	}

	// Approximate page count from PDF internals.
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
