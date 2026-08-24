package export

import (
	"testing"
	"time"
)

func TestSafeFileName(t *testing.T) {
	for _, tc := range []struct{ title, fallback, want string }{
		{"Anh Trai Nhân Vật Chính", "8476-x", "Anh-Trai-Nhân-Vật-Chính"},
		{"Tập 1: Mở Đầu / Kết", "x", "Tập-1-Mở-Đầu-Kết"},
		{"???", "8476-fallback", "8476-fallback"},
		{"", "8476-fallback", "8476-fallback"},
	} {
		if got := SafeFileName(tc.title, tc.fallback); got != tc.want {
			t.Errorf("SafeFileName(%q) = %q, want %q", tc.title, got, tc.want)
		}
	}
}

// Zero means "unset" for the tunables, so asking for no cache and no delay has
// to be said explicitly — otherwise the defaults come back.
func TestApplyDefaults(t *testing.T) {
	req := Request{NovelURL: "https://ln.hako.vn/sang-tac/1-x"}
	req.applyDefaults()

	if req.Page != DefaultPage || req.FontSize != DefaultFontSize || req.Workers != DefaultWorkers {
		t.Errorf("defaults not applied: %+v", req)
	}
	if req.CacheDir != DefaultCacheDir || req.Delay != DefaultDelay {
		t.Errorf("cache/delay defaults not applied: %q %v", req.CacheDir, req.Delay)
	}

	opted := Request{NovelURL: "https://ln.hako.vn/sang-tac/1-x", NoCache: true, NoDelay: true}
	opted.applyDefaults()
	if opted.CacheDir != "" {
		t.Errorf("NoCache should clear the cache dir, got %q", opted.CacheDir)
	}
	if opted.Delay != 0 {
		t.Errorf("NoDelay should clear the delay, got %v", opted.Delay)
	}
}

// validate runs before any request is made, so a bad flag costs no fetches.
func TestValidate(t *testing.T) {
	valid := "https://ln.hako.vn/sang-tac/1-x"
	for name, req := range map[string]Request{
		"no url":            {},
		"not http":          {NovelURL: "ftp://ln.hako.vn/x"},
		"unknown page":      {NovelURL: valid, Page: "billboard"},
		"zero workers":      {NovelURL: valid, Workers: -1},
		"negative margin":   {NovelURL: valid, Margin: -1},
		"negative fontsize": {NovelURL: valid, FontSize: -1},
	} {
		req.applyDefaults()
		if err := req.validate(); err == nil {
			t.Errorf("%s: expected an error, got none", name)
		}
	}

	ok := Request{NovelURL: valid, Delay: time.Second}
	ok.applyDefaults()
	if err := ok.validate(); err != nil {
		t.Errorf("valid request rejected: %v", err)
	}
}
