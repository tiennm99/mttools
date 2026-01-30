package main

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	_ "image/gif"  // Register GIF decoder
	_ "image/jpeg" // Register JPEG decoder
	_ "image/png"  // Register PNG decoder

	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
)

// ResizeImage resizes an image to a specified percentage of its original size
func ResizeImage(inputPath, outputPath string, scalePercent int) error {
	file, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("error decoding image: %w", err)
	}

	// Get original dimensions
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Calculate new dimensions
	newWidth := width * scalePercent / 100
	newHeight := height * scalePercent / 100

	// Create resized image using nearest neighbor (simple approach)
	// For better quality, consider using github.com/nfnt/resize
	resized := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	// Simple resampling
	for y := 0; y < newHeight; y++ {
		for x := 0; x < newWidth; x++ {
			srcX := x * width / newWidth
			srcY := y * height / newHeight
			resized.Set(x, y, img.At(srcX, srcY))
		}
	}

	// Create output file
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("error creating output file: %w", err)
	}
	defer outFile.Close()

	// Encode based on format or output extension
	ext := strings.ToLower(filepath.Ext(outputPath))
	switch ext {
	case ".jpg", ".jpeg":
		err = jpeg.Encode(outFile, resized, &jpeg.Options{Quality: 95})
	case ".png":
		err = png.Encode(outFile, resized)
	case ".gif":
		// For GIF, we'd need additional encoding support
		err = jpeg.Encode(outFile, resized, &jpeg.Options{Quality: 95})
	case ".bmp":
		err = bmp.Encode(outFile, resized)
	case ".tiff", ".tif":
		err = tiff.Encode(outFile, resized, nil)
	case ".webp":
		// WebP encoding is complex, save as PNG instead
		err = png.Encode(outFile, resized)
	default:
		// Default to PNG
		err = png.Encode(outFile, resized)
	}

	if err != nil {
		return fmt.Errorf("error encoding image: %w", err)
	}

	fmt.Printf("Resized: %s -> %s\n", inputPath, outputPath)
	fmt.Printf("  Original: %dx%d, New: %dx%d\n", width, height, newWidth, newHeight)

	return nil
}

// ResizeImagesInFolder resizes all images in a folder and its subfolders
func ResizeImagesInFolder(folderPath string, scalePercent int, outputFolder string, overwrite bool) error {
	// Supported image formats
	imageExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".bmp":  true,
		".tiff": true,
		".webp": true,
	}

	folderPath = filepath.Clean(folderPath)
	var outputBasePath string
	if overwrite {
		outputBasePath = folderPath
	} else {
		if outputFolder == "" {
			outputBasePath = filepath.Join(folderPath, "resized")
		} else {
			outputBasePath = filepath.Clean(outputFolder)
		}
	}

	// Walk through all files and subdirectories
	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if file has supported image extension
		ext := strings.ToLower(filepath.Ext(path))
		if !imageExtensions[ext] {
			return nil
		}

		// Determine output path
		var outputPath string
		if overwrite {
			outputPath = path
		} else {
			// Get relative path from folderPath
			relPath, err := filepath.Rel(folderPath, path)
			if err != nil {
				return fmt.Errorf("error getting relative path: %w", err)
			}
			outputPath = filepath.Join(outputBasePath, relPath)

			// Create output directory if it doesn't exist
			outputDir := filepath.Dir(outputPath)
			if err := os.MkdirAll(outputDir, 0755); err != nil {
				return fmt.Errorf("error creating output directory: %w", err)
			}
		}

		// Resize the image
		if err := ResizeImage(path, outputPath, scalePercent); err != nil {
			fmt.Printf("Error processing %s: %v\n", path, err)
		}

		return nil
	})

	return err
}

func main() {
	// Configuration
	folderPath := `D:\kvtm\kvtm\client\res\common\ui\EventTet2026` // Change to your folder path
	scalePercent := 50 // Resize to 50% of original size
	overwrite := true  // Set to false to save to output folder instead

	fmt.Printf("Image Resizer - Scaling to %d%%\n", scalePercent)
	fmt.Printf("Source folder: %s\n", folderPath)
	if overwrite {
		fmt.Println("Mode: Overwriting original images")
	} else {
		fmt.Println("Mode: Saving to output folder")
	}
	fmt.Println(strings.Repeat("-", 50))

	if err := ResizeImagesInFolder(folderPath, scalePercent, "", overwrite); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(strings.Repeat("-", 50))
	fmt.Println("Done!")
}
