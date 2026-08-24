package pdfout

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// bundledTTF is the font used when nothing else is available, so rendering never
// depends on the host having fonts installed — a minimal container, and the
// headless box this tool is usually run on, typically have none. See
// fonts/NOTICE.md for provenance and licensing.
//
//go:embed fonts/DejaVuSans.ttf
var bundledTTF []byte

// BundledFontName labels the embedded font in diagnostics. It is not a path;
// the font is compiled into the binary.
const BundledFontName = "DejaVu Sans (bundled)"

// Font is font data ready to embed in a PDF.
//
// The data is carried as bytes rather than as a path because fpdf joins a font
// path onto its own font directory, which defaults to "." — turning an absolute
// path into a working-directory-relative one that resolves only when the
// process happens to run from the filesystem root.
type Font struct {
	// Name identifies the font for diagnostics: the path it was read from, or
	// BundledFontName.
	Name string
	Data []byte
}

// Vietnamese text needs the Latin Extended Additional block (ư, ạ, ế, ộ …).
// Every font listed here ships with its platform and covers it; a basic-Latin
// font would silently drop the diacritics.
func fontCandidates() []string {
	switch runtime.GOOS {
	case "windows":
		dir := filepath.Join(os.Getenv("SystemRoot"), "Fonts")
		if os.Getenv("SystemRoot") == "" {
			dir = `C:\Windows\Fonts`
		}
		return []string{
			filepath.Join(dir, "segoeui.ttf"),
			filepath.Join(dir, "arial.ttf"),
			filepath.Join(dir, "calibri.ttf"),
			filepath.Join(dir, "tahoma.ttf"),
			filepath.Join(dir, "times.ttf"),
		}
	case "darwin":
		return []string{
			"/System/Library/Fonts/Supplemental/Arial.ttf",
			"/Library/Fonts/Arial.ttf",
			"/System/Library/Fonts/Supplemental/Times New Roman.ttf",
		}
	default:
		return []string{
			"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
			"/usr/share/fonts/truetype/noto/NotoSans-Regular.ttf",
			"/usr/share/fonts/TTF/DejaVuSans.ttf",
			"/usr/share/fonts/dejavu/DejaVuSans.ttf",
			"/usr/share/fonts/liberation/LiberationSans-Regular.ttf",
		}
	}
}

// FindFont returns the first available system font suitable for Vietnamese.
// Callers that just need something that works should use LoadFont, which falls
// back to the bundled font instead of failing.
func FindFont() (string, error) {
	candidates := fontCandidates()
	for _, path := range candidates {
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path, nil
		}
	}
	return "", fmt.Errorf("no Vietnamese-capable system font found (looked for %v)", candidates)
}

// LoadFont resolves the font to embed.
//
// An explicit path wins, and is a hard error when it cannot be read: a caller
// that named a font wants that font, not a silent substitute. Otherwise a
// system font is used, and when none can be read the bundled font is returned —
// so with an empty path LoadFont always succeeds.
func LoadFont(path string) (Font, error) {
	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return Font{}, fmt.Errorf("read font: %w", err)
		}
		return Font{Name: path, Data: data}, nil
	}
	if found, err := FindFont(); err == nil {
		if data, err := os.ReadFile(found); err == nil {
			return Font{Name: found, Data: data}, nil
		}
		// A listed font that cannot be read is no better than a missing one.
	}
	return BundledFont(), nil
}

// BundledFont returns the font compiled into the binary.
func BundledFont() Font {
	return Font{Name: BundledFontName, Data: bundledTTF}
}
