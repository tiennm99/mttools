package pdfout

import (
	"os"
	"path/filepath"
	"testing"
)

func testOptions(t *testing.T) Options {
	t.Helper()
	// LoadFont("") cannot fail — it ends at the bundled font — so unlike the
	// old FindFont call this never skips for want of a system font.
	font, err := LoadFont("")
	if err != nil {
		t.Fatalf("LoadFont: %v", err)
	}
	return Options{
		Page:        Presets["phone"],
		Margin:      6,
		Font:        font,
		FontSize:    12,
		LineSpacing: 1.55,
		Title:       "TRỞ LẠI NĂM THÁNG CŨ",
		SourceURL:   "https://example.test/n.html",
	}
}

func TestWriteProducesReadablePDF(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.pdf")

	chapters := []Chapter{
		{Heading: "Chương 1", Paragraphs: []string{
			"Nghe vị trưởng tử nói chuyện với nàng.",
			"Một đoạn văn khác để kiểm tra ngắt dòng tự động trên trang nhỏ.",
		}},
		{Heading: "Chương 2", Paragraphs: []string{"Đoạn cuối."}},
	}

	if err := Write(path, testOptions(t), chapters); err != nil {
		t.Fatalf("Write: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat output: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("wrote an empty PDF")
	}

	header := make([]byte, 5)
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if _, err := f.Read(header); err != nil {
		t.Fatal(err)
	}
	if string(header) != "%PDF-" {
		t.Errorf("output does not start with a PDF header, got %q", header)
	}
}

// A novel with many chapters must not overflow a page; auto page break plus the
// footer reserve handles that, so a long chapter should span several pages.
func TestWriteHandlesLongChapters(t *testing.T) {
	path := filepath.Join(t.TempDir(), "long.pdf")

	paragraphs := make([]string, 200)
	for i := range paragraphs {
		paragraphs[i] = "Một đoạn văn dài để buộc trình kết xuất phải sang trang mới nhiều lần."
	}

	if err := Write(path, testOptions(t), []Chapter{{Heading: "Chương 1", Paragraphs: paragraphs}}); err != nil {
		t.Fatalf("Write: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() < 2000 {
		t.Errorf("output suspiciously small (%d bytes) for 200 paragraphs", info.Size())
	}
}

func TestPresetNamesIsStable(t *testing.T) {
	got := PresetNames()
	want := []string{"a4", "a5", "phone"}
	if len(got) != len(want) {
		t.Fatalf("PresetNames() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("PresetNames()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
