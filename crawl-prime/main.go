package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/anaskhan96/soup"
)

const (
	targetURL    = "http://compoasso.free.fr/primelistweb/page/prime/liste_online_en.php"
	maxPrimes    = 1000000
	outputFile   = "primes.txt"
	perPageParam = "600"
)

func main() {
	current := 0
	totalWritten := 0

	for current < maxPrimes {
		fmt.Printf("Processing primes starting from: %d\n", current)

		primes, nextCurrent, err := fetchPrimesPage(current)
		if err != nil {
			log.Printf("Error fetching page starting at %d: %v", current, err)
			continue
		}

		if len(primes) == 0 {
			log.Printf("No primes found on page starting at %d", current)
			break
		}

		// Write primes to file
		err = writePrimesToFile(primes)
		if err != nil {
			log.Printf("Error writing primes to file: %v", err)
			continue
		}

		totalWritten += len(primes)
		fmt.Printf("Wrote %d primes (total: %d)\n", len(primes), totalWritten)

		if nextCurrent <= current {
			log.Printf("Next current (%d) is not greater than current (%d), stopping", nextCurrent, current)
			break
		}

		current = nextCurrent
	}

	fmt.Printf("Completed! Total primes written: %d\n", totalWritten)
}

func fetchPrimesPage(start int) ([]string, int, error) {
	// Prepare form data
	formData := url.Values{
		"primePageInput": {perPageParam},
		"numberInput":    {strconv.Itoa(start)},
	}

	// Create request
	req, err := http.NewRequest("POST", targetURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Prime-Crawler/1.0")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Parse HTML
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to read response body: %w", err)
	}

	return parsePrimesFromHTML(body, start)
}

func parsePrimesFromHTML(htmlContent []byte, start int) ([]string, int, error) {
	// Parse HTML using soup (similar to BeautifulSoup)
	htmlStr := string(htmlContent)
	doc := soup.HTMLParse(htmlStr)

	// Find the first table (matching Python soup.find('table'))
	table := doc.Find("table")
	if table.Error != nil {
		return nil, 0, fmt.Errorf("failed to find table: %w", table.Error)
	}

	// Find form within the table (matching Python table.find('form'))
	form := table.Find("form")
	if form.Error != nil {
		return nil, 0, fmt.Errorf("failed to find form: %w", form.Error)
	}

	// Find inner table within the form (matching Python form.find('table'))
	table1 := form.Find("table")
	if table1.Error != nil {
		return nil, 0, fmt.Errorf("failed to find inner table: %w", table1.Error)
	}

	// Find all td elements (matching Python table1.find_all('td'))
	tdElements := table1.FindAll("td")
	if len(tdElements) == 0 {
		return nil, 0, fmt.Errorf("no td elements found")
	}

	// Extract primes from all td elements except last 10 (matching Python td_elements[0:-10])
	var primes []string
	for i, td := range tdElements {
		if i < len(tdElements)-10 {
			text := strings.TrimSpace(td.Text())
			if text != "" {
				primes = append(primes, text)
			}
		}
	}

	// Get next current from the 11th element from the end (matching Python td_elements[-11])
	var nextCurrent int
	if len(tdElements) >= 11 {
		nextTd := tdElements[len(tdElements)-11]
		text := strings.TrimSpace(nextTd.Text())
		if text != "" {
			var err error
			nextCurrent, err = strconv.Atoi(text)
			if err != nil {
				return nil, 0, fmt.Errorf("failed to parse next current number: %w", err)
			}
			nextCurrent += 1
		}
	}

	if len(primes) == 0 {
		return nil, 0, fmt.Errorf("no primes found in HTML response")
	}

	return primes, nextCurrent, nil
}

func writePrimesToFile(primes []string) error {
	file, err := os.OpenFile(outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open output file: %w", err)
	}
	defer file.Close()

	for _, prime := range primes {
		_, err := file.WriteString(prime + "\n")
		if err != nil {
			return fmt.Errorf("failed to write prime to file: %w", err)
		}
	}

	return nil
}
