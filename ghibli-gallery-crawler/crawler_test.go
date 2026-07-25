package main

import (
	"net/url"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractImageURLs(t *testing.T) {
	base, err := url.Parse("https://example.com/gallery/index.html")
	if err != nil {
		t.Fatalf("parsing base: %v", err)
	}

	doc := `<html><body>
		<img src="photo.jpg">
		<img src="/absolute/other.png">
		<img src="https://cdn.example.net/remote.gif">
		<img src="/hsts-pixel.gif?c=3.2.5">
		<img src="photo.jpg">
		<img alt="no src attribute">
		<img src="data:image/png;base64,AAAA">
	</body></html>`

	got, err := extractImageURLs(base, strings.NewReader(doc))
	if err != nil {
		t.Fatalf("extractImageURLs: %v", err)
	}

	want := []string{
		"https://example.com/gallery/photo.jpg",
		"https://example.com/absolute/other.png",
		"https://cdn.example.net/remote.gif",
		"https://example.com/hsts-pixel.gif", // query stripped
	}

	if len(got) != len(want) {
		t.Fatalf("got %d URLs %v, want %d %v", len(got), got, len(want), want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("URL %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestGalleryJobs(t *testing.T) {
	jobs, err := galleryJobs("https://www.ghibli.jp/gallery/", []string{"ponyo", "ged"}, 3, "out")
	if err != nil {
		t.Fatalf("galleryJobs: %v", err)
	}

	if len(jobs) != 6 {
		t.Fatalf("got %d jobs, want 6", len(jobs))
	}

	// Numbers are zero-padded to three digits, as in the original script.
	if want := "https://www.ghibli.jp/gallery/ponyo001.jpg"; jobs[0].url != want {
		t.Errorf("first URL = %q, want %q", jobs[0].url, want)
	}
	if want := filepath.Join("out", "ponyo"); jobs[0].dir != want {
		t.Errorf("first dir = %q, want %q", jobs[0].dir, want)
	}
	if want := "https://www.ghibli.jp/gallery/ged003.jpg"; jobs[5].url != want {
		t.Errorf("last URL = %q, want %q", jobs[5].url, want)
	}
}

func TestFileNameFromURL(t *testing.T) {
	cases := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{in: "https://example.com/a/b/chihiro001.jpg", want: "chihiro001.jpg"},
		{in: "https://example.com/image.png", want: "image.png"},
		{in: "https://example.com/", wantErr: true},
		{in: "https://example.com", wantErr: true},
	}

	for _, c := range cases {
		got, err := fileNameFromURL(c.in)
		if c.wantErr {
			if err == nil {
				t.Errorf("fileNameFromURL(%q) = %q, want error", c.in, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("fileNameFromURL(%q): %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("fileNameFromURL(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestSplitFilms(t *testing.T) {
	got := splitFilms(" ponyo , ,ged, ../escape ")
	want := []string{"ponyo", "ged"}

	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("slug %d = %q, want %q", i, got[i], want[i])
		}
	}
}
