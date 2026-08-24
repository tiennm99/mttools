// Command hako-crawler downloads every chapter of an ln.hako.vn novel and
// exports it as a PDF sized for reading on a phone.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/tiennm99/hako-crawler/export"
	"github.com/tiennm99/hako-crawler/pdfout"
)

// exampleURL is the novel this project was originally written for; it stands in
// as the example now that any novel can be passed instead.
const exampleURL = "https://ln.hako.vn/sang-tac/8476-kiep-nay-la-anh-trai-cua-nhan-vat-chinh"

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {
	req, err := parseFlags()
	if err != nil {
		return err
	}

	// Ctrl-C cancels in-flight fetches instead of leaving a partial PDF.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	req.Log = func(format string, args ...any) {
		fmt.Fprintf(os.Stderr, format+"\n", args...)
	}

	result, err := export.Export(ctx, *req)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "\n%s\n", result.Summary())
	fmt.Fprintf(os.Stderr, "font: %s at %.0fpt on %s page (%.0f x %.0f mm)\n",
		filepath.Base(result.FontName), req.FontSize, result.Page.Name, result.Page.W, result.Page.H)
	if result.TextDir != "" {
		fmt.Fprintf(os.Stderr, "text: %s\n", result.TextDir)
	}
	fmt.Println(result.Path)
	return nil
}

// parseFlags builds the export request from the command line. Validation of the
// resulting values lives in export.Export, so the CLI and an embedding program
// reject the same inputs.
func parseFlags() (*export.Request, error) {
	req := &export.Request{}

	flag.StringVar(&req.NovelURL, "url", "", "novel page URL, e.g. "+exampleURL)
	flag.StringVar(&req.OutPath, "out", "", "output PDF path (default: the novel's name)")
	flag.StringVar(&req.TextDir, "txt", "", "also write one plain-text file per chapter into this directory")
	flag.StringVar(&req.Page, "page", export.DefaultPage,
		"page size: "+strings.Join(pdfout.PresetNames(), ", "))
	flag.StringVar(&req.FontFile, "font", "", "path to a .ttf font (default: a Vietnamese-capable system font)")
	flag.Float64Var(&req.FontSize, "font-size", export.DefaultFontSize, "body font size in points")
	flag.Float64Var(&req.LineSpacing, "line-spacing", export.DefaultLineSpacing, "line height as a multiple of font size")
	flag.Float64Var(&req.Margin, "margin", export.DefaultMargin, "page margin in millimetres")
	flag.IntVar(&req.Workers, "workers", export.DefaultWorkers, "concurrent chapter fetches")
	flag.DurationVar(&req.Delay, "delay", export.DefaultDelay, "minimum delay between requests")
	flag.IntVar(&req.Retries, "retries", export.DefaultRetries, "retries per request")
	flag.IntVar(&req.Limit, "limit", 0, "only fetch the first N chapters (0 = all)")
	flag.StringVar(&req.CacheDir, "cache", export.DefaultCacheDir,
		"directory for cached pages, so re-exports need no requests (empty to disable)")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(),
			"Download an ln.hako.vn novel and export it as a phone-friendly PDF.\n\n"+
				"Usage:\n  hako-crawler -url <novel page URL> [flags]\n\nFlags:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if req.NovelURL == "" {
		flag.Usage()
		return nil, fmt.Errorf("-url is required")
	}
	// The flag defaults above are already applied, so a zero value here can only
	// come from the user asking for none. Say so explicitly: left as the zero
	// value, Export would read it as "unset" and restore the default.
	req.NoCache = req.CacheDir == ""
	req.NoDelay = req.Delay == 0
	return req, nil
}
