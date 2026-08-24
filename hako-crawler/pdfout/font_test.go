package pdfout

import (
	"os"
	"path/filepath"
	"testing"
)

// With no path given, LoadFont must always succeed: it falls back to the
// bundled font so rendering never depends on the host having fonts installed.
func TestLoadFontFallsBackToBundled(t *testing.T) {
	font, err := LoadFont("")
	if err != nil {
		t.Fatalf("LoadFont(\"\"): %v", err)
	}
	if len(font.Data) == 0 {
		t.Error("font carries no data")
	}
	if font.Name == "" {
		t.Error("font has no name for diagnostics")
	}
}

// A named font that cannot be read is an error, not a silent substitution: a
// caller who asked for a font wants that font.
func TestLoadFontMissingPathIsAnError(t *testing.T) {
	if _, err := LoadFont(filepath.Join(t.TempDir(), "absent.ttf")); err == nil {
		t.Fatal("expected an error for an unreadable font path, got none")
	}
}

func TestLoadFontReadsGivenPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "given.ttf")
	if err := os.WriteFile(path, BundledFont().Data, 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	font, err := LoadFont(path)
	if err != nil {
		t.Fatalf("LoadFont(%q): %v", path, err)
	}
	if font.Name != path {
		t.Errorf("name = %q, want %q", font.Name, path)
	}
}

func TestBundledFontIsEmbedded(t *testing.T) {
	if len(BundledFont().Data) < 100_000 {
		t.Errorf("bundled font is %d bytes, which is too small to be a real TTF",
			len(BundledFont().Data))
	}
}
