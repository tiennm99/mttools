package export

import (
	"testing"
	"time"
)

func TestApplyDefaultsFillsUnsetFields(t *testing.T) {
	req := Request{NovelURL: "https://monkeydd.com/example.html"}
	req.applyDefaults()

	if req.Page != DefaultPage {
		t.Errorf("Page = %q, want %q", req.Page, DefaultPage)
	}
	if req.FontSize != DefaultFontSize {
		t.Errorf("FontSize = %v, want %v", req.FontSize, DefaultFontSize)
	}
	if req.LineSpacing != DefaultLineSpacing {
		t.Errorf("LineSpacing = %v, want %v", req.LineSpacing, DefaultLineSpacing)
	}
	if req.Margin != DefaultMargin {
		t.Errorf("Margin = %v, want %v", req.Margin, DefaultMargin)
	}
	if req.Workers != DefaultWorkers {
		t.Errorf("Workers = %v, want %v", req.Workers, DefaultWorkers)
	}
	if req.Delay != DefaultDelay {
		t.Errorf("Delay = %v, want %v", req.Delay, DefaultDelay)
	}
	if req.Retries != DefaultRetries {
		t.Errorf("Retries = %v, want %v", req.Retries, DefaultRetries)
	}
	if req.CacheDir != DefaultCacheDir {
		t.Errorf("CacheDir = %q, want %q", req.CacheDir, DefaultCacheDir)
	}
}

func TestApplyDefaultsKeepsExplicitValues(t *testing.T) {
	req := Request{
		NovelURL:    "https://monkeydd.com/example.html",
		Page:        "a5",
		FontSize:    14,
		LineSpacing: 1.4,
		Margin:      10,
		Workers:     2,
		Delay:       time.Second,
		Retries:     1,
		CacheDir:    "/tmp/pages",
	}
	req.applyDefaults()

	if req.Page != "a5" || req.FontSize != 14 || req.LineSpacing != 1.4 {
		t.Errorf("layout fields overwritten: %+v", req)
	}
	if req.Margin != 10 || req.Workers != 2 || req.Delay != time.Second || req.Retries != 1 {
		t.Errorf("fetch fields overwritten: %+v", req)
	}
	if req.CacheDir != "/tmp/pages" {
		t.Errorf("CacheDir = %q, want /tmp/pages", req.CacheDir)
	}
}

// A zero Margin is indistinguishable from "unset", so it becomes the default.
// NoCache and NoDelay exist precisely because "none" must survive that rule.
func TestApplyDefaultsHonoursOptOuts(t *testing.T) {
	req := Request{NovelURL: "https://monkeydd.com/example.html", NoCache: true, NoDelay: true}
	req.applyDefaults()

	if req.CacheDir != "" {
		t.Errorf("NoCache left CacheDir = %q, want empty", req.CacheDir)
	}
	if req.Delay != 0 {
		t.Errorf("NoDelay left Delay = %v, want 0", req.Delay)
	}
}

// Opt-outs win over an explicitly set value so a caller cannot end up with both.
func TestApplyDefaultsOptOutsOverrideExplicitValues(t *testing.T) {
	req := Request{
		NovelURL: "https://monkeydd.com/example.html",
		CacheDir: "/tmp/pages",
		NoCache:  true,
		Delay:    time.Second,
		NoDelay:  true,
	}
	req.applyDefaults()

	if req.CacheDir != "" {
		t.Errorf("CacheDir = %q, want empty", req.CacheDir)
	}
	if req.Delay != 0 {
		t.Errorf("Delay = %v, want 0", req.Delay)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		req     Request
		wantErr bool
	}{
		{
			name: "defaults are valid",
			req:  Request{NovelURL: "https://monkeydd.com/example.html"},
		},
		{
			name:    "missing url",
			req:     Request{},
			wantErr: true,
		},
		{
			name:    "non-http scheme",
			req:     Request{NovelURL: "ftp://monkeydd.com/example.html"},
			wantErr: true,
		},
		{
			name:    "scheme-less url",
			req:     Request{NovelURL: "monkeydd.com/example.html"},
			wantErr: true,
		},
		{
			name:    "unknown page preset",
			req:     Request{NovelURL: "https://monkeydd.com/example.html", Page: "letter"},
			wantErr: true,
		},
		{
			name:    "negative font size",
			req:     Request{NovelURL: "https://monkeydd.com/example.html", FontSize: -1},
			wantErr: true,
		},
		{
			name:    "negative line spacing",
			req:     Request{NovelURL: "https://monkeydd.com/example.html", LineSpacing: -1},
			wantErr: true,
		},
		{
			name:    "negative margin",
			req:     Request{NovelURL: "https://monkeydd.com/example.html", Margin: -1},
			wantErr: true,
		},
		{
			name:    "zero workers is filled by defaults, negative is not",
			req:     Request{NovelURL: "https://monkeydd.com/example.html", Workers: -1},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.req
			req.applyDefaults()
			err := req.validate()
			if tt.wantErr && err == nil {
				t.Error("validate() = nil, want error")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("validate() = %v, want nil", err)
			}
		})
	}
}

func TestSafeFileName(t *testing.T) {
	tests := []struct {
		title    string
		fallback string
		want     string
	}{
		{"Trở Lại Năm Tháng Cũ", "slug", "Trở-Lại-Năm-Tháng-Cũ"},
		{"Chapter: One / Two", "slug", "Chapter-One-Two"},
		{"  ---  ", "slug", "slug"},
		{"", "slug", "slug"},
	}
	for _, tt := range tests {
		if got := SafeFileName(tt.title, tt.fallback); got != tt.want {
			t.Errorf("SafeFileName(%q, %q) = %q, want %q", tt.title, tt.fallback, got, tt.want)
		}
	}
}

func TestResultSummary(t *testing.T) {
	r := &Result{Title: "Example", Chapters: 12, Words: 3400}
	want := "Example — 12 chapters, 3400 words"
	if got := r.Summary(); got != want {
		t.Errorf("Summary() = %q, want %q", got, want)
	}
}
