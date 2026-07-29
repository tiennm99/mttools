package pdfout

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// Vietnamese text needs the Latin Extended Additional block (ư, ạ, ế, ộ …).
// Every font listed here ships with its platform and covers it; fonts with only
// basic Latin would silently drop the diacritics.
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
func FindFont() (string, error) {
	candidates := fontCandidates()
	for _, path := range candidates {
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path, nil
		}
	}
	return "", fmt.Errorf("no Vietnamese-capable system font found (looked for %v); "+
		"pass -font with a path to a .ttf file", candidates)
}
