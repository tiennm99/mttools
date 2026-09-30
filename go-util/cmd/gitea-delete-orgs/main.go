package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	_ "github.com/tiennm99/mttools/go-util/internal/env"
)

// org represents a Gitea organization.
type org struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
}

func main() {
	baseURL := os.Getenv("GITEA_URL")
	token := os.Getenv("GITEA_TOKEN")
	keepList := os.Getenv("GITEA_KEEP_ORGS") // comma-separated org names to keep

	if baseURL == "" || token == "" {
		log.Fatal("GITEA_URL and GITEA_TOKEN env vars are required")
	}

	baseURL = strings.TrimRight(baseURL, "/")

	keep := make(map[string]bool)
	for _, name := range strings.Split(keepList, ",") {
		name = strings.TrimSpace(name)
		if name != "" {
			keep[name] = true
		}
	}

	orgs, err := listOrgs(baseURL, token)
	if err != nil {
		log.Fatalf("failed to list orgs: %v", err)
	}

	fmt.Printf("Found %d orgs, keeping: %v\n", len(orgs), mapsKeys(keep))

	for _, o := range orgs {
		if keep[o.Username] {
			fmt.Printf("  KEEP   %s\n", o.Username)
			continue
		}
		fmt.Printf("  DELETE %s ... ", o.Username)
		if err := deleteOrg(baseURL, token, o.Username); err != nil {
			fmt.Printf("FAILED: %v\n", err)
		} else {
			fmt.Println("OK")
		}
	}
}

// listOrgs fetches all organizations the authenticated user belongs to.
func listOrgs(baseURL, token string) ([]org, error) {
	var allOrgs []org
	page := 1

	for {
		url := fmt.Sprintf("%s/api/v1/user/orgs?page=%d&limit=50", baseURL, page)
		body, err := doRequest(http.MethodGet, url, token)
		if err != nil {
			return nil, err
		}

		var orgs []org
		if err := json.Unmarshal(body, &orgs); err != nil {
			return nil, fmt.Errorf("decode orgs: %w", err)
		}
		if len(orgs) == 0 {
			break
		}

		allOrgs = append(allOrgs, orgs...)
		page++
	}

	return allOrgs, nil
}

// deleteOrg deletes the organization entirely via DELETE /api/v1/orgs/{org}.
func deleteOrg(baseURL, token, orgName string) error {
	url := fmt.Sprintf("%s/api/v1/orgs/%s", baseURL, orgName)
	_, err := doRequest(http.MethodDelete, url, token)
	return err
}

// doRequest executes an HTTP request with token auth and returns the response body.
func doRequest(method, url, token string) ([]byte, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

func mapsKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
