package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sync"
)

// job is a single image to fetch and the directory to store it in.
type job struct {
	url string
	dir string
}

// errNotFound marks images the server does not have. The numbered gallery mode
// always asks for a fixed range, so missing images are expected, not failures.
var errNotFound = errors.New("not found")

// download fetches every job using at most concurrency parallel requests and
// reports how many succeeded. Individual failures are logged and do not stop
// the run.
func download(client *http.Client, jobs []job, concurrency int) error {
	if concurrency < 1 {
		concurrency = 1
	}
	if concurrency > len(jobs) {
		concurrency = len(jobs)
	}

	// Create directories up front so parallel workers never race on MkdirAll.
	for _, j := range jobs {
		if err := os.MkdirAll(j.dir, 0o755); err != nil {
			return fmt.Errorf("creating %s: %w", j.dir, err)
		}
	}

	var (
		queue   = make(chan job)
		wg      sync.WaitGroup
		mu      sync.Mutex
		saved   int
		missing int
		failed  int
	)

	for range concurrency {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range queue {
				size, err := downloadOne(client, j)

				mu.Lock()
				switch {
				case err == nil:
					saved++
					fmt.Printf("saved %s (%s)\n", j.url, humanSize(size))
				case errors.Is(err, errNotFound):
					missing++
				default:
					failed++
					fmt.Fprintf(os.Stderr, "failed %s: %v\n", j.url, err)
				}
				mu.Unlock()
			}
		}()
	}

	for _, j := range jobs {
		queue <- j
	}
	close(queue)
	wg.Wait()

	fmt.Printf("Done: %d saved, %d missing, %d failed\n", saved, missing, failed)
	if saved == 0 && failed > 0 {
		return fmt.Errorf("every download failed")
	}
	return nil
}

// downloadOne writes a single image to disk and returns the bytes written. An
// existing file with the same name is overwritten.
func downloadOne(client *http.Client, j job) (int64, error) {
	name, err := fileNameFromURL(j.url)
	if err != nil {
		return 0, err
	}

	resp, err := client.Get(j.url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusForbidden {
		return 0, errNotFound
	}
	if resp.StatusCode != http.StatusOK {
		// Bail out before writing, so error pages never land on disk as images.
		return 0, fmt.Errorf("unexpected status %s", resp.Status)
	}

	dest := filepath.Join(j.dir, name)
	f, err := os.Create(dest)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	written, err := io.Copy(f, resp.Body)
	if err != nil {
		// Leave no truncated file behind on a mid-transfer error.
		os.Remove(dest)
		return 0, err
	}
	return written, nil
}

// fileNameFromURL derives a safe local file name from the last path segment of
// an image URL.
func fileNameFromURL(rawURL string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	name := path.Base(parsed.Path)
	// path.Base returns "." or "/" when there is nothing usable to take.
	if name == "" || name == "." || name == "/" || name == ".." {
		return "", fmt.Errorf("cannot derive a file name from %q", rawURL)
	}
	// Guard against a remote path smuggling in separators.
	return filepath.Base(name), nil
}

// humanSize formats a byte count for display.
func humanSize(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	value := float64(n)
	for _, suffix := range []string{"KiB", "MiB", "GiB"} {
		value /= unit
		if value < unit {
			return fmt.Sprintf("%.1f %s", value, suffix)
		}
	}
	return fmt.Sprintf("%.1f TiB", value)
}
