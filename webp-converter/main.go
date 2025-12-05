package main

import (
	"bytes"
	"encoding/json"
	"image/png"
	"io"
	"net/http"

	// _ "image/gif"
	// _ "image/jpeg"
	// _ "image/png"

	"golang.org/x/image/webp"
)

type Request struct {
	Img string `json:"img"`
}

func main() {
	http.HandleFunc("/process", processHandler)
	http.ListenAndServe(":8080", nil)
}

func processHandler(w http.ResponseWriter, r *http.Request) {
	// Parse JSON
	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	resp, err := http.Get(req.Img)
	if err != nil {
		http.Error(w, "Failed to fetch image", http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, "Image URL returned non-200", http.StatusBadRequest)
		return
	}

	origBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read image", http.StatusInternalServerError)
		return
	}

	// Simple detection for WebP
	isWebP := bytes.Contains(origBytes[:32], []byte("WEBP"))

	if !isWebP {
		// Return the original image
		w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
		w.Write(origBytes)
		return
	}

	// Decode WebP
	img, err := webp.Decode(bytes.NewReader(origBytes))
	if err != nil {
		http.Error(w, "Failed to decode WebP", http.StatusInternalServerError)
		return
	}

	// Convert to PNG
	var out bytes.Buffer
	if err := png.Encode(&out, img); err != nil {
		http.Error(w, "Failed to encode PNG", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Write(out.Bytes())
}
