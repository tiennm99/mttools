package pdfout

import (
	"os"
	"path/filepath"
	"testing"
)

// Write must produce a real PDF with the bundled font, since the host this runs
// on has no fonts installed.
func TestWrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "book.pdf")

	opts := Options{
		Page:        Presets["phone"],
		Margin:      6,
		Font:        BundledFont(),
		FontSize:    10,
		LineSpacing: 1.55,
		Title:       "Truyện Thử",
		SourceURL:   "https://ln.hako.vn/sang-tac/1-x",
	}
	chapters := []Chapter{
		{Heading: "Chương 01", Volume: "Tập 01", Paragraphs: []string{"Trời hôm nay đẹp lắm."}},
		{Heading: "Chương 02", Volume: "Tập 01", Paragraphs: []string{"Người anh trai mỉm cười."}},
		{Heading: "Minh Họa 01", Volume: "Minh Họa", Paragraphs: nil},
	}
	if err := Write(path, opts, chapters); err != nil {
		t.Fatalf("write: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if len(data) < 1024 {
		t.Errorf("pdf is %d bytes, which is too small to hold an embedded font", len(data))
	}
	if string(data[:5]) != "%PDF-" {
		t.Errorf("missing PDF header, got %q", data[:5])
	}
}

func TestPresetNamesAreStable(t *testing.T) {
	want := []string{"a4", "a5", "phone"}
	got := PresetNames()
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}
