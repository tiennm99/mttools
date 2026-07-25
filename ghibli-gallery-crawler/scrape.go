package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"golang.org/x/net/html"
)

// scrapeImageURLs fetches a page and returns the absolute URLs of every image
// it references.
func scrapeImageURLs(client *http.Client, pageURL *url.URL) ([]string, error) {
	resp, err := client.Get(pageURL.String())
	if err != nil {
		return nil, fmt.Errorf("fetching %s: %w", pageURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching %s: unexpected status %s", pageURL, resp.Status)
	}

	return extractImageURLs(pageURL, resp.Body)
}

// extractImageURLs pulls the src of every <img> in the document, resolves it
// against base, and strips query strings so that cache-busting parameters do
// not end up in file names. Duplicates and non-absolute results are dropped.
func extractImageURLs(base *url.URL, r io.Reader) ([]string, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, fmt.Errorf("parsing HTML: %w", err)
	}

	var (
		urls []string
		seen = make(map[string]bool)
		walk func(*html.Node)
	)

	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "img" {
			if src, ok := attr(n, "src"); ok {
				if resolved, ok := resolveImageURL(base, src); ok && !seen[resolved] {
					seen[resolved] = true
					urls = append(urls, resolved)
				}
			}
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(doc)

	return urls, nil
}

// attr returns the value of the named attribute, if the node has it.
func attr(n *html.Node, name string) (string, bool) {
	for _, a := range n.Attr {
		if a.Key == name {
			return a.Val, true
		}
	}
	return "", false
}

// resolveImageURL makes src absolute relative to base and reports whether the
// result is a usable http(s) URL.
func resolveImageURL(base *url.URL, src string) (string, bool) {
	ref, err := url.Parse(src)
	if err != nil {
		return "", false
	}

	resolved := base.ResolveReference(ref)
	// Drop the query so URLs like '/hsts-pixel.gif?c=3.2.5' yield clean names.
	resolved.RawQuery = ""
	resolved.Fragment = ""

	if resolved.Host == "" {
		return "", false
	}
	// data: and other schemes are not downloadable files.
	if resolved.Scheme != "http" && resolved.Scheme != "https" {
		return "", false
	}
	return resolved.String(), true
}
