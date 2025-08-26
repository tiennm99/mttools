package main

import (
	"bufio"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
	"golang.org/x/net/html"
	"gopkg.in/yaml.v3"
)

type Config struct {
	RootURL   string   `yaml:"root_url"`
	Whitelist []string `yaml:"whitelist"`
	DataFile  string   `yaml:"data_file"`
	Delay     int      `yaml:"delay_seconds"`
}

type WebScraper struct {
	whitelist    []string
	processedURL map[string]bool
	mutex        sync.RWMutex
	dataFile     string
	delay        time.Duration
}

func NewWebScraper(whitelist []string, dataFile string, delay time.Duration) *WebScraper {
	return &WebScraper{
		whitelist:    whitelist,
		processedURL: make(map[string]bool),
		dataFile:     dataFile,
		delay:        delay,
	}
}

func (ws *WebScraper) loadProcessedURLs() error {
	file, err := os.Open(ws.dataFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		url := strings.TrimSpace(scanner.Text())
		if url != "" {
			ws.processedURL[url] = true
		}
	}
	return scanner.Err()
}

func (ws *WebScraper) saveProcessedURLs() error {
	file, err := os.Create(ws.dataFile)
	if err != nil {
		return err
	}
	defer file.Close()

	ws.mutex.RLock()
	defer ws.mutex.RUnlock()

	for url := range ws.processedURL {
		_, err := file.WriteString(url + "\n")
		if err != nil {
			return err
		}
	}
	return nil
}

func (ws *WebScraper) isWhitelisted(url string) bool {
	for _, pattern := range ws.whitelist {
		matched, err := regexp.MatchString(pattern, url)
		if err != nil {
			log.Printf("Invalid regex pattern '%s': %v", pattern, err)
			continue
		}
		if matched {
			return true
		}
	}
	return false
}

func (ws *WebScraper) isProcessed(url string) bool {
	ws.mutex.RLock()
	defer ws.mutex.RUnlock()
	return ws.processedURL[url]
}

func (ws *WebScraper) markAsProcessed(url string) {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()
	ws.processedURL[url] = true
}

func (ws *WebScraper) normalizeURLForFilename(urlStr string) string {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "invalid-url"
	}

	filename := parsedURL.Host + parsedURL.Path
	if parsedURL.RawQuery != "" {
		filename += "_" + parsedURL.RawQuery
	}

	reg := regexp.MustCompile(`[<>:"/\\|?*]`)
	filename = reg.ReplaceAllString(filename, "_")
	filename = strings.ReplaceAll(filename, ".", "_")

	if filename == "" || filename[len(filename)-1] == '_' {
		filename += "index"
	}

	return filename
}

func (ws *WebScraper) extractTextFromHTML(htmlContent string) string {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		log.Printf("Failed to parse HTML: %v", err)
		return htmlContent
	}

	var textContent strings.Builder
	ws.extractText(doc, &textContent)
	
	text := textContent.String()
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	text = strings.TrimSpace(text)
	
	return text
}

func (ws *WebScraper) extractText(n *html.Node, textContent *strings.Builder) {
	if n.Type == html.TextNode {
		textContent.WriteString(n.Data)
		textContent.WriteString(" ")
	}
	
	if n.Type == html.ElementNode {
		switch n.Data {
		case "script", "style", "noscript", "head":
			return
		case "br", "p", "div", "h1", "h2", "h3", "h4", "h5", "h6":
			textContent.WriteString(" ")
		}
	}
	
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		ws.extractText(child, textContent)
	}
}

func (ws *WebScraper) saveContent(urlStr, content string) error {
	dataDir := "data"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return err
	}

	filename := ws.normalizeURLForFilename(urlStr) + ".txt"
	filePath := filepath.Join(dataDir, filename)

	textContent := ws.extractTextFromHTML(content)
	return os.WriteFile(filePath, []byte(textContent), 0644)
}

func (ws *WebScraper) processURL(url string) {
	fmt.Printf("Processing URL: %s\n", url)
	ws.markAsProcessed(url)
}

func (ws *WebScraper) Start(rootURL string) error {
	if err := ws.loadProcessedURLs(); err != nil {
		return fmt.Errorf("failed to load processed URLs: %w", err)
	}

	c := colly.NewCollector(
		colly.AllowedDomains(),
	)

	c.OnResponse(func(r *colly.Response) {
		urlStr := r.Request.URL.String()
		if ws.isWhitelisted(urlStr) {
			content := string(r.Body)
			if err := ws.saveContent(urlStr, content); err != nil {
				log.Printf("Failed to save content for %s: %v", urlStr, err)
			} else {
				fmt.Printf("Saved content: %s\n", ws.normalizeURLForFilename(urlStr)+".txt")
			}
		}
	})

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		absoluteURL := e.Request.AbsoluteURL(link)

		if ws.isProcessed(absoluteURL) {
			return
		}

		if ws.isWhitelisted(absoluteURL) {
			ws.processURL(absoluteURL)
			c.Visit(absoluteURL)
		}
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Printf("Visiting: %s\n", r.URL.String())
		if ws.delay > 0 {
			time.Sleep(ws.delay)
		}
	})

	c.OnError(func(r *colly.Response, err error) {
		if r != nil {
			switch r.StatusCode {
			case 301, 302, 303, 307, 308:
				log.Printf("REDIRECT: %s (Status: %d) -> Location: %s", r.Request.URL, r.StatusCode, r.Headers.Get("Location"))
			case 403:
				log.Printf("BLOCKED: Access forbidden to %s (Status: 403)", r.Request.URL)
			case 429:
				log.Printf("RATE_LIMITED: Too many requests to %s (Status: 429)", r.Request.URL)
			case 404:
				log.Printf("NOT_FOUND: %s (Status: 404)", r.Request.URL)
			default:
				log.Printf("ERROR: %s (Status: %d) - %v", r.Request.URL, r.StatusCode, err)
			}
		} else {
			log.Printf("NETWORK_ERROR: Failed to connect to %s - %v", r.Request.URL, err)
		}
	})

	if ws.isWhitelisted(rootURL) && !ws.isProcessed(rootURL) {
		ws.processURL(rootURL)
	}

	err := c.Visit(rootURL)
	if err != nil {
		return fmt.Errorf("failed to visit root URL: %w", err)
	}

	if err := ws.saveProcessedURLs(); err != nil {
		return fmt.Errorf("failed to save processed URLs: %w", err)
	}

	fmt.Printf("Scraping completed. Processed %d URLs.\n", len(ws.processedURL))
	return nil
}

func loadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	return &config, err
}

func main() {
	config, err := loadConfig("config.yaml")
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	delay := time.Duration(config.Delay) * time.Second
	scraper := NewWebScraper(config.Whitelist, config.DataFile, delay)

	if err := scraper.Start(config.RootURL); err != nil {
		log.Fatal(err)
	}
}