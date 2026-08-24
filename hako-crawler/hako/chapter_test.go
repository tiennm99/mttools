package hako

import (
	"strings"
	"testing"
)

// chapterPage wraps a body in the page furniture the parser has to see past:
// the container, and the hidden copy of the chapter title the site plants
// beside it.
func chapterPage(body string) string {
	return `<html><body><div id="chapter-content" class="long-text text-justify">
	<p style="display: none">Chương 01</p>` + body + `</div></body></html>`
}

// protectedBody renders the encoded container the site now serves.
func protectedBody(t *testing.T, html, key string) string {
	t.Helper()
	chunks := encodeXORShuffle(t, html, key, 24)
	quoted := make([]string, len(chunks))
	for i, c := range chunks {
		// The attribute is HTML-escaped on the page; the parser decodes it.
		quoted[i] = "&quot;" + c + "&quot;"
	}
	return `<div id="` + protectedID + `" data-s="` + schemeXORShuffle +
		`" data-k="` + key + `" data-c="[` + strings.Join(quoted, ",") + `]" aria-hidden="true"></div>`
}

var testRef = ChapterRef{Label: "Chương 01", Volume: "Tập 01", URL: "https://ln.hako.vn/sang-tac/1-x/c1-chuong-01"}

func TestParseChapterProtected(t *testing.T) {
	body := `<p id="1">Chương 01</p>` +
		`<p id="2">Trời hôm nay đẹp lắm.</p>` +
		`<p id="3">Người anh trai <em>mỉm cười</em>.[note1]</p>` +
		`<p id="4">   </p>`

	ch, err := ParseChapter([]byte(chapterPage(protectedBody(t, body, testKey))), testRef)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	// The leading paragraph only repeats the label, the blank one carries no
	// text, and the footnote marker is page furniture rather than prose.
	want := []string{"Trời hôm nay đẹp lắm.", "Người anh trai mỉm cười."}
	if len(ch.Paragraphs) != len(want) {
		t.Fatalf("paragraphs = %q, want %q", ch.Paragraphs, want)
	}
	for i, w := range want {
		if ch.Paragraphs[i] != w {
			t.Errorf("paragraph %d = %q, want %q", i, ch.Paragraphs[i], w)
		}
	}
	if ch.Volume != testRef.Volume {
		t.Errorf("volume = %q, want %q", ch.Volume, testRef.Volume)
	}
	if ch.WordCount() != 10 {
		t.Errorf("word count = %d, want 10", ch.WordCount())
	}
}

// A chapter still served as plain markup must keep working, and the hidden
// title beside the body must not become a paragraph.
func TestParseChapterPlainMarkup(t *testing.T) {
	ch, err := ParseChapter([]byte(chapterPage(`<p id="1">Một buổi sáng yên tĩnh.</p>`)), testRef)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(ch.Paragraphs) != 1 || ch.Paragraphs[0] != "Một buổi sáng yên tĩnh." {
		t.Fatalf("paragraphs = %q", ch.Paragraphs)
	}
}

// Illustration chapters are real chapters that carry pictures and little or no
// text. They must not fail the run.
func TestParseChapterWithoutProse(t *testing.T) {
	ch, err := ParseChapter([]byte(chapterPage(protectedBody(t, `<p id="1"><img src="/a.png"></p>`, testKey))),
		ChapterRef{Label: "Minh Họa 01", URL: testRef.URL})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(ch.Paragraphs) != 0 {
		t.Errorf("paragraphs = %q, want none", ch.Paragraphs)
	}
}

func TestParseChapterMissingContainer(t *testing.T) {
	if _, err := ParseChapter([]byte(`<html><body><p>nope</p></body></html>`), testRef); err == nil {
		t.Fatal("expected an error when the content container is absent, got none")
	}
}

// A rotated scheme must surface as an error naming the chapter, not as an empty
// or garbled one.
func TestParseChapterUnknownScheme(t *testing.T) {
	page := chapterPage(`<div id="` + protectedID + `" data-s="aes_gcm_2027" data-k="k" data-c="[&quot;0000AAAA&quot;]"></div>`)
	_, err := ParseChapter([]byte(page), testRef)
	if err == nil {
		t.Fatal("expected an error for an unknown scheme, got none")
	}
	if !strings.Contains(err.Error(), testRef.URL) {
		t.Errorf("error should name the chapter URL, got: %v", err)
	}
}

func TestHeading(t *testing.T) {
	for _, tc := range []struct{ label, want string }{
		{"Chương 01", "Chương 01"},
		{"12", "Chương 12"},
		{"", "Chương"},
		{"Minh Họa 01", "Minh Họa 01"},
	} {
		if got := (&Chapter{Label: tc.label}).Heading(); got != tc.want {
			t.Errorf("Heading(%q) = %q, want %q", tc.label, got, tc.want)
		}
	}
}
