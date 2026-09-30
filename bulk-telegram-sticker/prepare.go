package main

import (
	"encoding/json"
	"fmt"
	"image"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/chai2010/webp"
	"github.com/disintegration/imaging"
)

type Sticker struct {
	Original string `json:"original"`
	Webp     string `json:"webp"`
	Emoji    string `json:"emoji"`
}

type Config struct {
	Title     string    `json:"title"`
	ShortName string    `json:"short_name"`
	Stickers  []Sticker `json:"stickers"`
}

func convertToWebp(inputPath, outputPath string, size image.Point) error {
	srcImg, err := imaging.Open(inputPath)
	if err != nil {
		return fmt.Errorf("open %s: %w", inputPath, err)
	}

	// Convert to RGBA
	rgba := imaging.ToRGBA(srcImg)

	// Resize maintaining aspect ratio
	resized := imaging.Fit(rgba, size.X, size.Y, imaging.Lanczos)

	// Pad to exact size with transparent background
	dst := image.NewRGBA(size)
	offset := imaging.Center(size.X, size.Y, resized.Bounds().Dx(), resized.Bounds().Dy())
	drawer := imaging.NewDrawer(dst)
	drawer.Draw(offset.X, offset.Y, resized)

	// Save as WebP
	opts := &webp.Options{Lossless: false, Quality: 90}
	if err := webp.EncodeFile(outputPath, dst, opts); err != nil {
		return fmt.Errorf("encode webp %s: %w", outputPath, err)
	}
	return nil
}

func processImages(inputDir, outputDir string) ([]Sticker, error) {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, err
	}

	var processed []Sticker
	err := filepath.WalkDir(inputDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".png" && ext != ".jpg" && ext != ".jpeg" {
			return nil
		}

		filename := d.Name()
		nameWithoutExt := strings.TrimSuffix(filename, filepath.Ext(filename))
		webpPath := filepath.Join(outputDir, nameWithoutExt+".webp")

		if err := convertToWebp(path, webpPath, image.Pt(512, 512)); err != nil {
			fmt.Printf("Error converting %s: %v\n", filename, err)
			return nil
		}

		processed = append(processed, Sticker{
			Original: filename,
			Webp:     nameWithoutExt + ".webp",
			Emoji:    "😀", // Default
		})
		fmt.Printf("Processed %s -> %s\n", filename, nameWithoutExt+".webp")
		return nil
	})
	return processed, err
}

func main() {
	inputDir := "input"
	outputDir := "output"

	if err := os.MkdirAll(inputDir, 0755); err != nil {
		fmt.Printf("Error creating input dir: %v\n", err)
		return
	}

	fmt.Println("Processing images...")
	stickers, err := processImages(inputDir, outputDir)
	if err != nil {
		fmt.Printf("Error processing images: %v\n", err)
		return
	}
	if len(stickers) == 0 {
		fmt.Println("No images found in input directory!")
		return
	}

	config := Config{
		Title:     "My Awesome Stickers",
		ShortName: "my_awesome_stickers",
		Stickers:  stickers,
	}

	configPath := filepath.Join(outputDir, "sticker_config.json")
	data, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		fmt.Printf("Error marshaling config: %v\n", err)
		return
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		fmt.Printf("Error writing config: %v\n", err)
		return
	}

	fmt.Printf("Processed %d images\n", len(stickers))
	fmt.Printf("Configuration generated: %s\n", configPath)
	fmt.Println("\nEdit sticker_config.json to customize:")
	fmt.Println("1. Change title and short_name")
	fmt.Println("2. Set custom emojis for stickers")
	fmt.Println("3. Add/remove stickers")
	fmt.Println("\nThen run: go run main.go")
}
