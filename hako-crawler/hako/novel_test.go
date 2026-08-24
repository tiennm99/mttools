package hako

import "testing"

// A landing page reduced to the structure the parser depends on: a series name,
// genre links (including one hidden behind the site's "show more" toggle), and
// two volume sections whose chapters are listed in reading order.
const novelFixture = `<!doctype html><html><head>
<title>Truyện Thử - Cổng Light Novel - Đọc Light Novel</title></head><body>
<div class="series-name-group"><span class="series-name">
  <a href="/sang-tac/1234-truyen-thu">Truyện Thử</a></span></div>
<div class="series-gernes">
  <a class="series-gerne-item" href="/the-loai/comedy">Comedy</a>
  <a class="series-gerne-item" href="/the-loai/fantasy">Fantasy</a>
  <a class="series-gerne-item" style="display: none;" href="/the-loai/magic">Magic</a>
  <a class="series-gerne-item" href="/the-loai/comedy">Comedy</a>
</div>
<section class="volume-list at-series basic-section">
  <header id="volume_1" class="sect-header"><span class="sect-title"> Minh Họa </span></header>
  <main><ul>
    <li><div class="chapter-name"><i class="fas fa-image"></i>
      <a href="/sang-tac/1234-truyen-thu/c11-minh-hoa-01" title="Minh Họa 01">Minh Họa 01</a>
    </div><div class="chapter-time">01/01/2024</div></li>
  </ul></main>
</section>
<section class="volume-list at-series basic-section">
  <header id="volume_2" class="sect-header"><span class="sect-title">Tập 01</span></header>
  <main><ul>
    <li><div class="chapter-name">
      <a href="/sang-tac/1234-truyen-thu/c21-chuong-01" title="Chương 01">Chương 01</a></div></li>
    <li><div class="chapter-name">
      <a href="https://ln.hako.vn/sang-tac/1234-truyen-thu/c22-chuong-02" title="Chương 02">Chương 02</a></div></li>
    <li><div class="chapter-name">
      <a href="/sang-tac/1234-truyen-thu/c21-chuong-01" title="Chương 01">Chương 01</a></div></li>
  </ul></main>
</section>
</body></html>`

const novelURL = "https://ln.hako.vn/sang-tac/1234-truyen-thu"

func TestParseNovelPage(t *testing.T) {
	novel, err := ParseNovelPage([]byte(novelFixture), novelURL)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if novel.Title != "Truyện Thử" {
		t.Errorf("title = %q, want %q", novel.Title, "Truyện Thử")
	}
	if novel.Slug != "1234-truyen-thu" {
		t.Errorf("slug = %q, want %q", novel.Slug, "1234-truyen-thu")
	}

	// The hidden genre is still this novel's genre; the repeated one is not a
	// second genre.
	wantTags := []string{"Comedy", "Fantasy", "Magic"}
	if len(novel.Tags) != len(wantTags) {
		t.Fatalf("tags = %v, want %v", novel.Tags, wantTags)
	}
	for i, want := range wantTags {
		if novel.Tags[i] != want {
			t.Errorf("tag %d = %q, want %q", i, novel.Tags[i], want)
		}
	}

	// Document order is reading order, volume by volume, and the duplicated
	// link must not produce a duplicated chapter.
	want := []ChapterRef{
		{Label: "Minh Họa 01", Volume: "Minh Họa", URL: novelURL + "/c11-minh-hoa-01"},
		{Label: "Chương 01", Volume: "Tập 01", URL: novelURL + "/c21-chuong-01"},
		{Label: "Chương 02", Volume: "Tập 01", URL: novelURL + "/c22-chuong-02"},
	}
	if len(novel.Chapters) != len(want) {
		t.Fatalf("got %d chapters, want %d: %+v", len(novel.Chapters), len(want), novel.Chapters)
	}
	for i, w := range want {
		if novel.Chapters[i] != w {
			t.Errorf("chapter %d = %+v, want %+v", i, novel.Chapters[i], w)
		}
	}
}

// A layout with no volume sections must still yield its chapters rather than
// reporting an empty novel.
func TestParseNovelPageWithoutVolumes(t *testing.T) {
	page := `<html><body><span class="series-name"><a href="/x">Truyện Thử</a></span>
	<div class="chapter-name"><a href="/sang-tac/1234-truyen-thu/c1-chuong-01">Chương 01</a></div>
	</body></html>`

	novel, err := ParseNovelPage([]byte(page), novelURL)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(novel.Chapters) != 1 {
		t.Fatalf("got %d chapters, want 1", len(novel.Chapters))
	}
	if novel.Chapters[0].Volume != "" {
		t.Errorf("volume = %q, want empty", novel.Chapters[0].Volume)
	}
}

func TestParseNovelPageWithNoChapters(t *testing.T) {
	page := `<html><body><span class="series-name"><a href="/x">Truyện Thử</a></span></body></html>`
	if _, err := ParseNovelPage([]byte(page), novelURL); err == nil {
		t.Fatal("expected an error when the page lists no chapters, got none")
	}
}

// Without a series heading the document title is the only source, and it
// carries the site's name as a suffix that must not end up in the file name.
func TestParseNovelPageTitleFallback(t *testing.T) {
	page := `<html><head><title>Truyện Thử - Cổng Light Novel - Đọc Light Novel</title></head>
	<body><div class="chapter-name"><a href="/c1">Chương 01</a></div></body></html>`

	novel, err := ParseNovelPage([]byte(page), novelURL)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if novel.Title != "Truyện Thử" {
		t.Errorf("title = %q, want %q", novel.Title, "Truyện Thử")
	}
}
