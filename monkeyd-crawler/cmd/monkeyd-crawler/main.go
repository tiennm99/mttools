// Command monkeyd-crawler downloads every chapter of a monkeydd.com novel and
// exports it as a PDF sized for reading on a phone.
package main

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/tiennm99/monkeyd-crawler/internal/monkeyd"
	"github.com/tiennm99/monkeyd-crawler/internal/pdfout"
)

type config struct {
	novelURL    string
	out         string
	page        string
	fontFile    string
	fontSize    float64
	lineSpacing float64
	margin      float64
	workers     int
	delay       time.Duration
	retries     int
	limit       int
	cacheDir    string
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := parseFlags()
	if err != nil {
		return err
	}

	// Ctrl-C cancels in-flight fetches instead of leaving a partial PDF.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	crawler := &monkeyd.Crawler{
		Client:   monkeyd.NewClient(cfg.delay, cfg.retries),
		CacheDir: cfg.cacheDir,
		Workers:  cfg.workers,
		Log: func(format string, args ...any) {
			fmt.Fprintf(os.Stderr, format+"\n", args...)
		},
	}

	novel, err := crawler.Novel(ctx, cfg.novelURL)
	if err != nil {
		return err
	}

	if cfg.limit > 0 && cfg.limit < len(novel.Chapters) {
		fmt.Fprintf(os.Stderr, "limiting to first %d of %d chapters\n", cfg.limit, len(novel.Chapters))
		novel.Chapters = novel.Chapters[:cfg.limit]
	}

	chapters, err := crawler.Chapters(ctx, novel)
	if err != nil {
		return err
	}

	fontFile := cfg.fontFile
	if fontFile == "" {
		if fontFile, err = pdfout.FindFont(); err != nil {
			return err
		}
	}

	outPath := cfg.out
	if outPath == "" {
		outPath = safeFileName(novel.Title, novel.Slug) + ".pdf"
	}

	opts := pdfout.Options{
		Page:        pdfout.Presets[cfg.page],
		Margin:      cfg.margin,
		FontFile:    fontFile,
		FontSize:    cfg.fontSize,
		LineSpacing: cfg.lineSpacing,
		Title:       novel.Title,
		SourceURL:   novel.URL,
	}
	if err := pdfout.Write(outPath, opts, toPDFChapters(chapters)); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "\n%s\n", monkeyd.Describe(novel, chapters))
	fmt.Fprintf(os.Stderr, "font: %s at %.0fpt on %s page (%.0f x %.0f mm)\n",
		filepath.Base(fontFile), cfg.fontSize, opts.Page.Name, opts.Page.W, opts.Page.H)
	fmt.Println(outPath)
	return nil
}

func parseFlags() (*config, error) {
	cfg := &config{}

	flag.StringVar(&cfg.novelURL, "url", "", "novel page URL, e.g. https://monkeydd.com/tro-lai-nam-thang-cu.html")
	flag.StringVar(&cfg.out, "out", "", "output PDF path (default: novel title)")
	flag.StringVar(&cfg.page, "page", "phone",
		"page size: "+strings.Join(pdfout.PresetNames(), ", "))
	flag.StringVar(&cfg.fontFile, "font", "", "path to a .ttf font (default: a Vietnamese-capable system font)")
	flag.Float64Var(&cfg.fontSize, "font-size", 12, "body font size in points")
	flag.Float64Var(&cfg.lineSpacing, "line-spacing", 1.55, "line height as a multiple of font size")
	flag.Float64Var(&cfg.margin, "margin", 6, "page margin in millimetres")
	flag.IntVar(&cfg.workers, "workers", 4, "concurrent chapter fetches")
	flag.DurationVar(&cfg.delay, "delay", 400*time.Millisecond, "minimum delay between requests")
	flag.IntVar(&cfg.retries, "retries", 3, "retries per request")
	flag.IntVar(&cfg.limit, "limit", 0, "only fetch the first N chapters (0 = all)")
	flag.StringVar(&cfg.cacheDir, "cache", ".cache",
		"directory for cached pages, so re-exports need no requests (empty to disable)")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(),
			"Download a monkeydd.com novel and export it as a phone-friendly PDF.\n\n"+
				"Usage:\n  monkeyd-crawler -url <novel page URL> [flags]\n\nFlags:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if cfg.novelURL == "" {
		flag.Usage()
		return nil, fmt.Errorf("-url is required")
	}
	parsed, err := url.Parse(cfg.novelURL)
	if err != nil {
		return nil, fmt.Errorf("invalid -url: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, fmt.Errorf("invalid -url: want an http(s) URL, got %q", cfg.novelURL)
	}
	if _, ok := pdfout.Presets[cfg.page]; !ok {
		return nil, fmt.Errorf("unknown -page %q: want one of %s",
			cfg.page, strings.Join(pdfout.PresetNames(), ", "))
	}
	if cfg.fontSize <= 0 {
		return nil, fmt.Errorf("-font-size must be positive")
	}
	if cfg.lineSpacing <= 0 {
		return nil, fmt.Errorf("-line-spacing must be positive")
	}
	if cfg.margin < 0 {
		return nil, fmt.Errorf("-margin cannot be negative")
	}
	if cfg.workers < 1 {
		return nil, fmt.Errorf("-workers must be at least 1")
	}
	return cfg, nil
}

func toPDFChapters(chapters []*monkeyd.Chapter) []pdfout.Chapter {
	out := make([]pdfout.Chapter, 0, len(chapters))
	for _, ch := range chapters {
		out = append(out, pdfout.Chapter{Heading: ch.Heading(), Paragraphs: ch.Paragraphs})
	}
	return out
}

var unsafeNameChars = regexp.MustCompile(`[^\p{L}\p{N}]+`)

// safeFileName builds a file name from the novel title, falling back to the
// slug when the title has no usable characters.
func safeFileName(title, fallback string) string {
	name := strings.Trim(unsafeNameChars.ReplaceAllString(title, "-"), "-")
	if name == "" {
		return fallback
	}
	return name
}
