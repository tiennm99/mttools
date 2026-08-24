package hako

import (
	"bytes"
	"fmt"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// ChapterRef points at one chapter listed on a novel page.
type ChapterRef struct {
	Label  string // as shown on the site, e.g. "Chương 01"
	Volume string // the volume section it sits under; empty if ungrouped
	URL    string
}

// Novel is a novel's landing page: what it is called and its chapters in
// reading order.
type Novel struct {
	Title string
	Slug  string
	URL   string

	// Tags are the novel's genres as labelled on the site, in page order.
	Tags     []string
	Chapters []ChapterRef
}

// ParseNovelPage reads the title, genres and chapter list off a novel page.
//
// Chapter URLs always come from the anchors on the page. Hako's chapter slugs
// carry an opaque id ("c75189-chuong-01") that cannot be derived from a chapter
// number, so a generated URL would simply 404.
func ParseNovelPage(page []byte, pageURL string) (*Novel, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(page))
	if err != nil {
		return nil, fmt.Errorf("parse novel page: %w", err)
	}
	base, err := url.Parse(pageURL)
	if err != nil {
		return nil, fmt.Errorf("parse novel url: %w", err)
	}

	novel := &Novel{
		Title:    novelTitle(doc),
		Slug:     slugFromURL(base),
		URL:      pageURL,
		Tags:     novelTags(doc),
		Chapters: chapterRefs(doc, base),
	}
	if novel.Title == "" {
		novel.Title = novel.Slug
	}
	if len(novel.Chapters) == 0 {
		return nil, fmt.Errorf("no chapters found on %s (page layout may have changed)", pageURL)
	}
	return novel, nil
}

// novelTitle prefers the series heading and falls back to the document title,
// which carries the site's name as a suffix and so is only a last resort.
func novelTitle(doc *goquery.Document) string {
	if t := collapseSpaces(doc.Find(".series-name a").First().Text()); t != "" {
		return t
	}
	if t := collapseSpaces(doc.Find(".series-name").First().Text()); t != "" {
		return t
	}
	title := collapseSpaces(doc.Find("title").First().Text())
	if head, _, found := strings.Cut(title, " - "); found {
		return head
	}
	return title
}

// novelTags reads the genre links out of the series info block.
//
// The class spelling is the site's own ("gerne"). Some entries are rendered
// hidden behind a "show more" toggle; they are still this novel's genres, so
// they are kept.
func novelTags(doc *goquery.Document) []string {
	var tags []string
	seen := make(map[string]bool)
	doc.Find("a.series-gerne-item").Each(func(_ int, s *goquery.Selection) {
		name := collapseSpaces(s.Text())
		if name == "" || seen[name] {
			return
		}
		seen[name] = true
		tags = append(tags, name)
	})
	return tags
}

// slugFromURL turns /sang-tac/8476-kiep-nay-la-anh-trai-cua-nhan-vat-chinh
// into "8476-kiep-nay-la-anh-trai-cua-nhan-vat-chinh".
func slugFromURL(u *url.URL) string {
	seg := strings.Trim(u.Path, "/")
	if i := strings.LastIndex(seg, "/"); i >= 0 {
		seg = seg[i+1:]
	}
	return seg
}

// chapterRefs collects every chapter link, tagged with the volume it sits
// under. The site lists volumes oldest first and chapters ascending within a
// volume, so document order is already reading order.
func chapterRefs(doc *goquery.Document, base *url.URL) []ChapterRef {
	var refs []ChapterRef
	seen := make(map[string]bool)

	add := func(volume string, sel *goquery.Selection) {
		sel.Find("div.chapter-name a").Each(func(_ int, a *goquery.Selection) {
			href, ok := a.Attr("href")
			if !ok || strings.TrimSpace(href) == "" {
				return
			}
			abs, err := base.Parse(strings.TrimSpace(href))
			if err != nil {
				return
			}
			link := abs.String()
			if seen[link] {
				return
			}
			seen[link] = true

			// The visible text is the chapter name; the title attribute
			// repeats it and survives when the text is decorated with icons.
			label := collapseSpaces(a.Text())
			if label == "" {
				label = collapseSpaces(a.AttrOr("title", ""))
			}
			refs = append(refs, ChapterRef{Label: label, Volume: volume, URL: link})
		})
	}

	volumes := doc.Find(".volume-list")
	volumes.Each(func(_ int, vol *goquery.Selection) {
		add(collapseSpaces(vol.Find("header .sect-title").First().Text()), vol)
	})
	// A layout without volume sections still has chapter links; picking them up
	// from the document keeps the crawl working rather than reporting an empty
	// novel. Links already taken from a volume are skipped by seen.
	add("", doc.Selection)

	return refs
}

// collapseSpaces trims and reduces every whitespace run — including the
// non-breaking spaces the site emits as &nbsp; — to a single space.
func collapseSpaces(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
