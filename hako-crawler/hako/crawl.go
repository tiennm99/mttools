package hako

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"
)

// Crawler fetches a novel and its chapters.
type Crawler struct {
	Client *Client

	// CacheDir, when set, stores raw pages on disk and serves later runs from
	// them, so re-rendering a book with different page or font settings costs
	// no requests.
	CacheDir string

	// Workers bounds concurrent fetches. The client's delay still caps the
	// overall request rate, so this trades latency, not politeness.
	Workers int

	// Log receives progress messages. Optional.
	Log func(format string, args ...any)
}

func (c *Crawler) logf(format string, args ...any) {
	if c.Log != nil {
		c.Log(format, args...)
	}
}

// Novel fetches a novel landing page and reads its chapter list.
func (c *Crawler) Novel(ctx context.Context, novelURL string) (*Novel, error) {
	page, err := c.page(ctx, novelURL)
	if err != nil {
		return nil, err
	}
	novel, err := ParseNovelPage(page, novelURL)
	if err != nil {
		return nil, err
	}
	c.logf("novel: %s (%d chapters)", novel.Title, len(novel.Chapters))
	return novel, nil
}

// Chapters fetches every chapter concurrently and returns them in reading
// order. A chapter that cannot be fetched or decoded fails the whole run
// rather than yielding a book with a silent hole in it.
func (c *Crawler) Chapters(ctx context.Context, novel *Novel) ([]*Chapter, error) {
	chapters := make([]*Chapter, len(novel.Chapters))

	workers := c.Workers
	if workers < 1 {
		workers = 1
	}

	group, groupCtx := errgroup.WithContext(ctx)
	group.SetLimit(workers)

	var mu sync.Mutex
	done := 0

	for i, ref := range novel.Chapters {
		group.Go(func() error {
			page, err := c.page(groupCtx, ref.URL)
			if err != nil {
				return err
			}
			chapter, err := ParseChapter(page, ref)
			if err != nil {
				return err
			}
			chapters[i] = chapter

			mu.Lock()
			done++
			c.logf("fetched %d/%d: %s (%d words)", done, len(novel.Chapters),
				chapter.Heading(), chapter.WordCount())
			mu.Unlock()
			return nil
		})
	}

	if err := group.Wait(); err != nil {
		return nil, err
	}
	return chapters, nil
}

// page returns a page from the cache when it is there, otherwise fetches and
// caches it.
func (c *Crawler) page(ctx context.Context, pageURL string) ([]byte, error) {
	path := c.cachePath(pageURL)
	if path != "" {
		if body, err := os.ReadFile(path); err == nil && len(body) > 0 {
			return body, nil
		}
	}

	body, err := c.Client.Get(ctx, pageURL)
	if err != nil {
		return nil, err
	}

	if path != "" {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err == nil {
			// A cache that cannot be written must not fail the crawl.
			_ = os.WriteFile(path, body, 0o644)
		}
	}
	return body, nil
}

// unsafeFileChars matches everything not allowed in a cache file name.
var unsafeFileChars = regexp.MustCompile(`[^A-Za-z0-9._-]+`)

// cachePath maps a page URL to a cache file, or "" when caching is off.
func (c *Crawler) cachePath(pageURL string) string {
	if c.CacheDir == "" {
		return ""
	}
	u, err := url.Parse(pageURL)
	if err != nil {
		return ""
	}
	name := unsafeFileChars.ReplaceAllString(strings.Trim(u.Path, "/"), "_")
	if name == "" {
		return ""
	}
	return filepath.Join(c.CacheDir, name+".html")
}

// TotalWords sums the word count across chapters.
func TotalWords(chapters []*Chapter) int {
	n := 0
	for _, ch := range chapters {
		n += ch.WordCount()
	}
	return n
}

// Describe renders a one-line summary of a crawl result.
func Describe(novel *Novel, chapters []*Chapter) string {
	return fmt.Sprintf("%s — %d chapters, %d words", novel.Title, len(chapters), TotalWords(chapters))
}
