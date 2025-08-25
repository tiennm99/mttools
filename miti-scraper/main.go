package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/gocolly/colly/v2"
)

type Config struct {
	RootURL   string   `json:"root_url"`
	Whitelist []string `json:"whitelist"`
	DataFile  string   `json:"data_file"`
}

type WebScraper struct {
	whitelist    []string
	processedURL map[string]bool
	mutex        sync.RWMutex
	dataFile     string
}

func NewWebScraper(whitelist []string, dataFile string) *WebScraper {
	return &WebScraper{
		whitelist:    whitelist,
		processedURL: make(map[string]bool),
		dataFile:     dataFile,
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
		if strings.Contains(url, pattern) {
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
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Printf("Error visiting %s: %v", r.Request.URL, err)
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
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	return &config, err
}

func main() {
	config, err := loadConfig("config.json")
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	scraper := NewWebScraper(config.Whitelist, config.DataFile)

	if err := scraper.Start(config.RootURL); err != nil {
		log.Fatal(err)
	}
}