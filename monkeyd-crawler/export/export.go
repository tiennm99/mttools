// Package export turns a novel URL into a PDF file. It holds the sequence the
// CLI and any embedding program both need — resolve the chapter list, fetch the
// chapters, pick a font, render the PDF — so neither has to reassemble it.
package export

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/tiennm99/monkeyd-crawler/monkeyd"
	"github.com/tiennm99/monkeyd-crawler/pdfout"
)

// Defaults for every tunable field of Request. Exported so a caller's own
// flags or config can advertise the same values instead of restating them.
const (
	DefaultPage        = "phone"
	DefaultFontSize    = 12.0
	DefaultLineSpacing = 1.55
	DefaultMargin      = 6.0
	DefaultWorkers     = 4
	DefaultDelay       = 400 * time.Millisecond
	DefaultRetries     = 3
	DefaultCacheDir    = ".cache"
)

// Request describes one export. Only NovelURL is required; every zero-valued
// tunable falls back to its Default above, so a caller that only has a URL can
// leave the rest alone.
type Request struct {
	NovelURL string

	// OutPath is the exact PDF path to write. When empty the file is named
	// after the novel title and placed in OutDir.
	OutPath string
	OutDir  string

	Page        string  // preset key: phone, a5, a4
	FontFile    string  // path to a .ttf; empty means discover a system font
	FontSize    float64 // points
	LineSpacing float64 // multiple of font size
	Margin      float64 // millimetres

	Workers int
	Retries int

	// Delay is the minimum spacing between requests. Set NoDelay to remove
	// the spacing rather than setting this to 0, which is read as "unset"
	// and gets the default back.
	Delay   time.Duration
	NoDelay bool

	// Limit caps the export to the first N chapters. 0 means every chapter.
	Limit int

	// CacheDir stores raw pages so a re-export costs no requests. Set
	// NoCache to opt out rather than clearing this field, which would be
	// read as "unset" and get the default back.
	CacheDir string
	NoCache  bool

	// Log receives progress messages. Optional.
	Log func(format string, args ...any)
}

// Result reports what was produced.
type Result struct {
	Path      string
	Title     string
	SourceURL string
	Chapters  int
	Words     int
	FontFile  string
	Page      pdfout.PageSize
}

// Summary renders a one-line description of the exported book.
func (r *Result) Summary() string {
	return fmt.Sprintf("%s — %d chapters, %d words", r.Title, r.Chapters, r.Words)
}

// Export fetches the novel at req.NovelURL and writes it as a PDF, returning
// where it landed. The context bounds the whole crawl; cancelling it abandons
// the run without leaving a partial PDF behind.
func Export(ctx context.Context, req Request) (*Result, error) {
	req.applyDefaults()
	if err := req.validate(); err != nil {
		return nil, err
	}

	crawler := &monkeyd.Crawler{
		Client:   monkeyd.NewClient(req.Delay, req.Retries),
		CacheDir: req.CacheDir,
		Workers:  req.Workers,
		Log:      req.Log,
	}

	novel, err := crawler.Novel(ctx, req.NovelURL)
	if err != nil {
		return nil, err
	}

	if req.Limit > 0 && req.Limit < len(novel.Chapters) {
		req.logf("limiting to first %d of %d chapters", req.Limit, len(novel.Chapters))
		novel.Chapters = novel.Chapters[:req.Limit]
	}

	chapters, err := crawler.Chapters(ctx, novel)
	if err != nil {
		return nil, err
	}

	fontFile := req.FontFile
	if fontFile == "" {
		if fontFile, err = pdfout.FindFont(); err != nil {
			return nil, err
		}
	}

	outPath := req.OutPath
	if outPath == "" {
		outPath = filepath.Join(req.OutDir, SafeFileName(novel.Title, novel.Slug)+".pdf")
	}

	page := pdfout.Presets[req.Page]
	opts := pdfout.Options{
		Page:        page,
		Margin:      req.Margin,
		FontFile:    fontFile,
		FontSize:    req.FontSize,
		LineSpacing: req.LineSpacing,
		Title:       novel.Title,
		SourceURL:   novel.URL,
	}
	if err := pdfout.Write(outPath, opts, toPDFChapters(chapters)); err != nil {
		return nil, err
	}

	return &Result{
		Path:      outPath,
		Title:     novel.Title,
		SourceURL: novel.URL,
		Chapters:  len(chapters),
		Words:     monkeyd.TotalWords(chapters),
		FontFile:  fontFile,
		Page:      page,
	}, nil
}

func (r *Request) applyDefaults() {
	if r.Page == "" {
		r.Page = DefaultPage
	}
	if r.FontSize == 0 {
		r.FontSize = DefaultFontSize
	}
	if r.LineSpacing == 0 {
		r.LineSpacing = DefaultLineSpacing
	}
	if r.Margin == 0 {
		r.Margin = DefaultMargin
	}
	if r.Workers == 0 {
		r.Workers = DefaultWorkers
	}
	switch {
	case r.NoDelay:
		r.Delay = 0
	case r.Delay == 0:
		r.Delay = DefaultDelay
	}
	if r.Retries == 0 {
		r.Retries = DefaultRetries
	}
	switch {
	case r.NoCache:
		r.CacheDir = ""
	case r.CacheDir == "":
		r.CacheDir = DefaultCacheDir
	}
}

// validate rejects a Request before any request is made, so a typo costs no
// fetches. Call applyDefaults first: it checks the effective values.
func (r *Request) validate() error {
	if r.NovelURL == "" {
		return fmt.Errorf("novel url is required")
	}
	parsed, err := url.Parse(r.NovelURL)
	if err != nil {
		return fmt.Errorf("invalid novel url: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("invalid novel url: want an http(s) URL, got %q", r.NovelURL)
	}
	if _, ok := pdfout.Presets[r.Page]; !ok {
		return fmt.Errorf("unknown page %q: want one of %s",
			r.Page, strings.Join(pdfout.PresetNames(), ", "))
	}
	if r.FontSize <= 0 {
		return fmt.Errorf("font size must be positive")
	}
	if r.LineSpacing <= 0 {
		return fmt.Errorf("line spacing must be positive")
	}
	if r.Margin < 0 {
		return fmt.Errorf("margin cannot be negative")
	}
	if r.Workers < 1 {
		return fmt.Errorf("workers must be at least 1")
	}
	return nil
}

func (r *Request) logf(format string, args ...any) {
	if r.Log != nil {
		r.Log(format, args...)
	}
}

func toPDFChapters(chapters []*monkeyd.Chapter) []pdfout.Chapter {
	out := make([]pdfout.Chapter, 0, len(chapters))
	for _, ch := range chapters {
		out = append(out, pdfout.Chapter{Heading: ch.Heading(), Paragraphs: ch.Paragraphs})
	}
	return out
}

var unsafeNameChars = regexp.MustCompile(`[^\p{L}\p{N}]+`)

// SafeFileName builds a file name from the novel title, falling back to the
// slug when the title has no usable characters. The result has no extension.
func SafeFileName(title, fallback string) string {
	name := strings.Trim(unsafeNameChars.ReplaceAllString(title, "-"), "-")
	if name == "" {
		return fallback
	}
	return name
}
