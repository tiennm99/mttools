// Command ghibli-gallery-crawler downloads images from web pages.
//
// It offers two modes, mirroring the two scripts it replaces:
//
//	scrape  - parse a page's <img> tags and download every image found
//	gallery - walk the numbered per-film galleries on ghibli.jp
package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	// defaultGalleryBase is where the numbered film galleries live. The
	// original script used http://; https:// reaches the same files.
	defaultGalleryBase = "https://www.ghibli.jp/gallery/"
	// defaultGalleryCount is the highest image number tried per film.
	defaultGalleryCount = 50
)

// defaultFilms lists the gallery slugs the original JavaScript script crawled.
var defaultFilms = []string{
	"marnie", "kaguyahime", "kazetachinu", "kokurikozaka",
	"karigurashi", "ponyo", "ged", "chihiro",
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		usage()
		return fmt.Errorf("no command given")
	}

	switch args[0] {
	case "scrape":
		return runScrape(args[1:])
	case "gallery":
		return runGallery(args[1:])
	case "help", "-h", "--help":
		usage()
		return nil
	default:
		usage()
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `ghibli-gallery-crawler downloads images from web pages.

Usage:
  ghibli-gallery-crawler scrape <url> [flags]
  ghibli-gallery-crawler gallery [flags]

Commands:
  scrape    Download every <img> found on a single page.
  gallery   Download the numbered ghibli.jp galleries for a list of films.

Run a command with -h to see its flags.
`)
}

// runScrape downloads every image referenced by a single page. The default
// output directory is the page's host, matching the Python script it replaces.
func runScrape(args []string) error {
	fs := flag.NewFlagSet("scrape", flag.ExitOnError)
	path := fs.String("path", "", "directory to store images in (default: the URL's host)")
	concurrency := fs.Int("concurrency", 8, "number of parallel downloads")
	timeout := fs.Duration("timeout", 30*time.Second, "per-request timeout")
	userAgent := fs.String("user-agent", defaultUserAgent, "User-Agent header to send")
	fs.Usage = func() {
		fmt.Fprintln(fs.Output(), "Usage: ghibli-gallery-crawler scrape <url> [flags]")
		fs.PrintDefaults()
	}
	positional, err := parseWithPositionals(fs, args)
	if err != nil {
		return err
	}

	if len(positional) != 1 {
		fs.Usage()
		return fmt.Errorf("scrape needs exactly one URL argument, got %d", len(positional))
	}
	pageURL := positional[0]

	parsed, err := url.Parse(pageURL)
	if err != nil {
		return fmt.Errorf("parsing %q: %w", pageURL, err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return fmt.Errorf("%q is not an absolute URL", pageURL)
	}

	dir := *path
	if dir == "" {
		dir = parsed.Host
	}

	client := newClient(*timeout, *userAgent)

	imageURLs, err := scrapeImageURLs(client, parsed)
	if err != nil {
		return err
	}
	if len(imageURLs) == 0 {
		fmt.Printf("No images found on %s\n", pageURL)
		return nil
	}
	fmt.Printf("Found %d image(s) on %s\n", len(imageURLs), pageURL)

	jobs := make([]job, 0, len(imageURLs))
	for _, imageURL := range imageURLs {
		jobs = append(jobs, job{url: imageURL, dir: dir})
	}
	return download(client, jobs, *concurrency)
}

// runGallery downloads images named <film>001.jpg .. <film>NNN.jpg for each
// film, storing each film's images in its own directory. This mirrors the
// JavaScript script it replaces.
func runGallery(args []string) error {
	fs := flag.NewFlagSet("gallery", flag.ExitOnError)
	base := fs.String("base", defaultGalleryBase, "base gallery URL")
	films := fs.String("films", strings.Join(defaultFilms, ","), "comma-separated film slugs")
	count := fs.Int("count", defaultGalleryCount, "highest image number to try per film")
	out := fs.String("out", ".", "directory to create the per-film directories in")
	concurrency := fs.Int("concurrency", 8, "number of parallel downloads")
	timeout := fs.Duration("timeout", 30*time.Second, "per-request timeout")
	userAgent := fs.String("user-agent", defaultUserAgent, "User-Agent header to send")
	fs.Usage = func() {
		fmt.Fprintln(fs.Output(), "Usage: ghibli-gallery-crawler gallery [flags]")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *count < 1 {
		return fmt.Errorf("-count must be at least 1, got %d", *count)
	}

	slugs := splitFilms(*films)
	if len(slugs) == 0 {
		return fmt.Errorf("-films must name at least one film")
	}

	jobs, err := galleryJobs(*base, slugs, *count, *out)
	if err != nil {
		return err
	}
	fmt.Printf("Trying %d image(s) across %d film(s)\n", len(jobs), len(slugs))

	client := newClient(*timeout, *userAgent)
	return download(client, jobs, *concurrency)
}

// parseWithPositionals parses flags that may appear before or after positional
// arguments. The standard flag package stops at the first non-flag argument, so
// the remainder is fed back through the parser one positional at a time.
func parseWithPositionals(fs *flag.FlagSet, args []string) ([]string, error) {
	var positional []string
	for len(args) > 0 {
		if err := fs.Parse(args); err != nil {
			return nil, err
		}
		args = fs.Args()
		if len(args) > 0 {
			positional = append(positional, args[0])
			args = args[1:]
		}
	}
	return positional, nil
}

// splitFilms turns a comma-separated flag value into clean slugs, dropping
// empty entries and anything that would escape the output directory.
func splitFilms(value string) []string {
	var slugs []string
	for _, raw := range strings.Split(value, ",") {
		slug := strings.TrimSpace(raw)
		if slug == "" {
			continue
		}
		// A slug becomes a directory name, so reject path separators.
		if slug != filepath.Base(slug) || slug == "." || slug == ".." {
			fmt.Fprintf(os.Stderr, "skipping invalid film slug %q\n", slug)
			continue
		}
		slugs = append(slugs, slug)
	}
	return slugs
}
