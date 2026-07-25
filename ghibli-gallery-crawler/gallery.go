package main

import (
	"fmt"
	"net/url"
	"path/filepath"
)

// galleryJobs builds the download list for the numbered ghibli.jp galleries:
// for each film, images are named <film>001.jpg through <film><count>.jpg and
// are stored in a directory named after the film.
func galleryJobs(base string, films []string, count int, out string) ([]job, error) {
	baseURL, err := url.Parse(base)
	if err != nil {
		return nil, fmt.Errorf("parsing base URL %q: %w", base, err)
	}
	if baseURL.Scheme == "" || baseURL.Host == "" {
		return nil, fmt.Errorf("base URL %q must be absolute", base)
	}

	jobs := make([]job, 0, len(films)*count)
	for _, film := range films {
		dir := filepath.Join(out, film)
		for i := 1; i <= count; i++ {
			name := fmt.Sprintf("%s%03d.jpg", film, i)
			imageURL, err := url.JoinPath(base, name)
			if err != nil {
				return nil, fmt.Errorf("building URL for %s: %w", name, err)
			}
			jobs = append(jobs, job{url: imageURL, dir: dir})
		}
	}
	return jobs, nil
}
