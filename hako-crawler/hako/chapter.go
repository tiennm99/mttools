package hako

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// contentID is the container holding a chapter's body.
const contentID = "chapter-content"

// Chapter is one fetched chapter reduced to plain paragraphs.
type Chapter struct {
	Label      string
	Volume     string
	URL        string
	Paragraphs []string
}

// Heading is the chapter title to print. A bare number reads poorly as a
// heading, so those get the Vietnamese word for "chapter".
func (c *Chapter) Heading() string {
	label := strings.TrimSpace(c.Label)
	if label == "" {
		return "Chương"
	}
	if isDigits(label) {
		return "Chương " + label
	}
	return label
}

// WordCount is a rough word count, used to sanity-check extraction.
func (c *Chapter) WordCount() int {
	n := 0
	for _, p := range c.Paragraphs {
		n += len(strings.Fields(p))
	}
	return n
}

func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// noteMarker matches the footnote placeholders left in the body text. The site
// swaps each one for a tooltip icon whose text lives elsewhere on the page, so
// in a plain-text export the marker itself is noise.
var noteMarker = regexp.MustCompile(`(?i)\[note\d+\]`)

// ParseChapter extracts a chapter's paragraphs.
//
// The body is normally delivered as an encoded payload rather than as markup
// (see protected.go); when it is, the payload is decoded first and the
// paragraphs are read out of the result. A chapter still served as plain markup
// is read directly, so both layouts work.
func ParseChapter(page []byte, ref ChapterRef) (*Chapter, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(page))
	if err != nil {
		return nil, fmt.Errorf("parse chapter %s: %w", ref.URL, err)
	}
	content := doc.Find("#" + contentID).First()
	if content.Length() == 0 {
		return nil, fmt.Errorf("chapter %s: no #%s container (page layout may have changed)",
			ref.URL, contentID)
	}

	body, err := chapterBody(content, ref)
	if err != nil {
		return nil, err
	}

	return &Chapter{
		Label:      ref.Label,
		Volume:     ref.Volume,
		URL:        ref.URL,
		Paragraphs: dropRepeatedTitle(extractParagraphs(body), ref.Label),
	}, nil
}

// chapterBody returns the selection actually holding the prose: the decoded
// payload when the body is encoded, otherwise the container itself.
func chapterBody(content *goquery.Selection, ref ChapterRef) (*goquery.Selection, error) {
	protected := content.Find("#" + protectedID).First()
	if protected.Length() == 0 {
		return content, nil
	}

	load, err := parsePayload(
		protected.AttrOr("data-s", ""),
		protected.AttrOr("data-k", ""),
		protected.AttrOr("data-c", "[]"),
	)
	if err != nil {
		return nil, fmt.Errorf("chapter %s: %w", ref.URL, err)
	}
	html, err := load.decode()
	if err != nil {
		return nil, fmt.Errorf("chapter %s: %w", ref.URL, err)
	}

	// The payload is a fragment of the chapter body only, with none of the
	// page's own furniture around it.
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("chapter %s: parse decoded body: %w", ref.URL, err)
	}
	return doc.Selection, nil
}

// extractParagraphs reads the paragraph text out of a body selection.
func extractParagraphs(body *goquery.Selection) []string {
	var paragraphs []string
	body.Find("p").Each(func(_ int, p *goquery.Selection) {
		// The page plants a hidden copy of the chapter title next to the body.
		if strings.Contains(strings.ReplaceAll(p.AttrOr("style", ""), " ", ""), "display:none") {
			return
		}
		text := collapseSpaces(noteMarker.ReplaceAllString(p.Text(), ""))
		if text == "" {
			return
		}
		paragraphs = append(paragraphs, text)
	})
	return paragraphs
}

// dropRepeatedTitle removes a leading paragraph that only repeats the chapter
// label, since the export prints its own heading.
func dropRepeatedTitle(paragraphs []string, label string) []string {
	if len(paragraphs) < 2 {
		return paragraphs
	}
	if strings.EqualFold(strings.TrimSpace(paragraphs[0]), strings.TrimSpace(label)) {
		return paragraphs[1:]
	}
	return paragraphs
}
