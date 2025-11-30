package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const DEBUG = false

var sanitizeRe = regexp.MustCompile(`[\\\/:*?"<>|]`)

func isDigit(id string) bool {
	if id == "" {
		return false
	}
	for _, r := range id {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func getFromURL(url string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP error: %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func writeTextToFile(text, filename string) error {
	filename = sanitizeRe.ReplaceAllString(filename, "_")
	dir := "data"
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	path := filepath.Join(dir, filename)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(text)
	return err
}

func readTextFromFile(filename string) (string, error) {
	path := filepath.Join("data", filename)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func main() {
	var html string
	var err error
	if DEBUG {
		html, err = readTextFromFile("_.txt")
	} else {
		html, err = getFromURL("https://ln.hako.vn/sang-tac/8476-kiep-nay-la-anh-trai-cua-nhan-vat-chinh")
		// writeTextToFile(html, "_.txt")
	}
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		fmt.Printf("Parse error: %v\n", err)
		return
	}

	chapters := doc.Find("div.chapter-name")
	/*
	   you may get `HTTP Error 429: Too Many Requests`
	   you can try again later, and skip downloaded chapters
	   example skip 50 first chapters like: chapters = chapters.Slice(50, chapters.Length())
	*/
	// chapters = chapters.Slice(100, chapters.Length())
	for i := 0; i < chapters.Length(); i++ {
		chapterSel := chapters.Eq(i)
		children := chapterSel.ChildrenFiltered("a")
		if children.Length() == 0 {
			continue
		}
		linkSel := children.First()
		title, hasTitle := linkSel.Attr("title")
		if !hasTitle {
			continue
		}
		href, hasHref := linkSel.Attr("href")
		if !hasHref {
			continue
		}
		chapterURL := "https://ln.hako.vn" + href
		chapterHTML, err := getFromURL(chapterURL)
		if err != nil {
			fmt.Printf("Error fetching %s: %v\n", chapterURL, err)
			continue
		}
		chapterDoc, err := goquery.NewDocumentFromReader(strings.NewReader(chapterHTML))
		if err != nil {
			fmt.Printf("Parse chapter %s: %v\n", title, err)
			continue
		}
		content := chapterDoc.Find("#chapter-content")
		var chapterData strings.Builder
		content.Find("p").Each(func(_ int, s *goquery.Selection) {
			id, exists := s.Attr("id")
			if exists && isDigit(id) {
				chapterData.WriteString(s.Text())
				chapterData.WriteString("\n")
			}
		})
		filename := title + ".txt"
		if err := writeTextToFile(chapterData.String(), filename); err != nil {
			fmt.Printf("Error writing %s: %v\n", filename, err)
		}
	}
}
