package pdfout

import (
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/image/font/sfnt"
)

func TestLoadFontUsesExplicitPath(t *testing.T) {
	// Any readable file is enough: LoadFont does not parse, it only reads.
	path := filepath.Join(t.TempDir(), "custom.ttf")
	want := []byte("not really a font")
	if err := os.WriteFile(path, want, 0o600); err != nil {
		t.Fatal(err)
	}

	font, err := LoadFont(path)
	if err != nil {
		t.Fatalf("LoadFont: %v", err)
	}
	if font.Name != path {
		t.Errorf("Name = %q, want %q", font.Name, path)
	}
	if string(font.Data) != string(want) {
		t.Errorf("Data = %q, want %q", font.Data, want)
	}
}

// A caller that names a font wants that font; silently substituting another
// would produce a PDF that does not match the request.
func TestLoadFontFailsOnUnreadableExplicitPath(t *testing.T) {
	font, err := LoadFont(filepath.Join(t.TempDir(), "missing.ttf"))
	if err == nil {
		t.Fatalf("LoadFont = %+v, want an error", font)
	}
}

// With no path given, LoadFont must always produce something usable — that is
// the whole point of the bundled font.
func TestLoadFontFallsBackWithoutExplicitPath(t *testing.T) {
	font, err := LoadFont("")
	if err != nil {
		t.Fatalf("LoadFont(\"\") = %v, want no error", err)
	}
	if len(font.Data) == 0 {
		t.Error("resolved font carries no data")
	}
	if font.Name == "" {
		t.Error("resolved font has no name")
	}
	// On a host with fonts installed this is a system path; on one without, it
	// is the bundled font. Either is fine, but it must be one of them.
	if system, err := FindFont(); err == nil {
		if font.Name != system && font.Name != BundledFontName {
			t.Errorf("Name = %q, want %q or %q", font.Name, system, BundledFontName)
		}
	} else if font.Name != BundledFontName {
		t.Errorf("Name = %q, want %q when no system font exists", font.Name, BundledFontName)
	}
}

func TestBundledFontIsParseable(t *testing.T) {
	font := BundledFont()
	if font.Name != BundledFontName {
		t.Errorf("Name = %q, want %q", font.Name, BundledFontName)
	}
	if _, err := sfnt.Parse(font.Data); err != nil {
		t.Fatalf("bundled font does not parse as a TrueType font: %v", err)
	}
}

// The bundled font exists to render Vietnamese. A replacement that lacks these
// glyphs would silently emit blanks, so check before trusting it.
func TestBundledFontCoversVietnamese(t *testing.T) {
	parsed, err := sfnt.Parse(BundledFont().Data)
	if err != nil {
		t.Fatalf("parse bundled font: %v", err)
	}
	var buf sfnt.Buffer
	for _, r := range []rune{'ư', 'ơ', 'đ', 'ạ', 'ế', 'ộ', 'ữ', 'ằ', 'ỷ', 'ỹ', 'Ọ', 'Ế', 'Ư', 'Đ'} {
		index, err := parsed.GlyphIndex(&buf, r)
		if err != nil {
			t.Errorf("GlyphIndex(%q): %v", r, err)
			continue
		}
		// Glyph 0 is .notdef — the character is absent from the font.
		if index == 0 {
			t.Errorf("bundled font has no glyph for %q (U+%04X)", r, r)
		}
	}
}
