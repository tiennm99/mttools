# monkeyd-crawler

Downloads every chapter of a [monkeydd.com](https://monkeydd.com) novel and exports it as a
single PDF sized for reading on a phone.

## Install

Requires Go 1.22+.

```sh
go build ./cmd/monkeyd-crawler
```

## Usage

Pass the novel's landing page URL:

```sh
./monkeyd-crawler -url https://monkeydd.com/tro-lai-nam-thang-cu.html
```

The PDF is named after the novel unless you pass `-out`.

```sh
# Bigger type, A5 page for a tablet
./monkeyd-crawler -url https://monkeydd.com/tro-lai-nam-thang-cu.html \
  -page a5 -font-size 14 -out truyen.pdf

# Try the layout on 3 chapters before fetching the whole book
./monkeyd-crawler -url https://monkeydd.com/tro-lai-nam-thang-cu.html -limit 3
```

### Flags

| Flag | Default | Purpose |
| --- | --- | --- |
| `-url` | *required* | Novel landing page URL |
| `-out` | novel title | Output PDF path |
| `-page` | `phone` | Page size: `phone`, `a5`, `a4` |
| `-font-size` | `12` | Body font size in points |
| `-line-spacing` | `1.55` | Line height as a multiple of font size |
| `-margin` | `6` | Page margin in mm |
| `-font` | auto | Path to a `.ttf`; defaults to a system font with Vietnamese coverage |
| `-workers` | `4` | Concurrent chapter fetches |
| `-delay` | `400ms` | Minimum delay between requests |
| `-retries` | `3` | Retries per request |
| `-limit` | `0` | Fetch only the first N chapters (0 = all) |
| `-cache` | `.cache` | Cache directory for raw pages (empty to disable) |

## Why the default page is 90×160 mm

Phone readability is governed by page *shape* more than by font size. A PDF viewer scales a
whole page to fit the screen, so a large font on an A4 page still ends up small: the page is
about three times wider than a phone screen and gets shrunk to match. The default page is cut
to a 9:16 ratio so it fills the screen at 100% zoom, where 12 pt renders at a comfortable
size with roughly 35–40 characters per line.

Use `-page a5` or `-page a4` for a tablet or for printing.

## How it works

1. Fetch the landing page and read the chapter list from `div.list-chapters`.
2. Cross-check that list against the `#selected_chapter` dropdown embedded in the first
   chapter page. When the dropdown is a superset it wins; a disagreement is reported.
3. Fetch each chapter concurrently and extract its text.
4. Render one PDF, each chapter starting on a new page.

### Two site behaviours the extractor has to handle

**Chapter text is partly served through CSS.** The markup contains empty elements, and the
stylesheet supplies the missing word:

```html
Nghe <span class="t-3e625e…"></span> trưởng tử
```

```css
.t-3e625e…:before { content: "vị"; }
```

Reading DOM text alone silently drops these — about 19% of the words in a sampled chapter. The
extractor parses the `:before` rules and substitutes each word back in.

**Chapter URLs cannot be generated.** Numbering has gaps (the sample novel has no chapter 4)
and slugs are not uniform across novels (`/14.html` on one, `/chuong-12.html` on another), so
chapter links are always parsed from the page rather than constructed from a count.

## Politeness and caching

Requests are spaced by `-delay` globally, so raising `-workers` does not raise the request
rate. Raw pages are cached under `.cache/`, so re-exporting with different font or page
settings costs no requests. Delete the directory to refetch.

## Tests

```sh
go test ./...
```

Tests run against synthetic fixtures that reproduce the CSS-injected words, the newest-first
chapter ordering, and the numbering gap. No network access required.

## Layout

```
cmd/monkeyd-crawler/    CLI
export/                 URL -> PDF in one call; shared by the CLI and importers
monkeyd/                fetching, HTML/CSS parsing, crawl orchestration
pdfout/                 PDF rendering and font discovery
```

## Use as a library

The packages are importable, so another Go program can produce the same PDF
without shelling out to the binary. `export.Export` is the whole pipeline —
chapter list, fetch, font discovery, render:

```go
result, err := export.Export(ctx, export.Request{
	NovelURL: "https://monkeydd.com/tro-lai-nam-thang-cu.html",
	OutDir:   tmpDir,
})
```

Only `NovelURL` is required; each zero-valued field falls back to the same
default as the matching CLI flag. Because zero means "unset", ask for *no*
cache or *no* request delay with the `NoCache` and `NoDelay` fields rather than
by zeroing `CacheDir` or `Delay`. Pass a `Log` function to receive the progress
messages the CLI prints to stderr.

Callers running in a container should note that `pdfout.FindFont` searches
system font paths: a minimal image with no fonts installed needs either a font
present or an explicit `FontFile`.

## Scope

Downloaded text stays on your machine; only fetch content you are allowed to read offline, and
respect the site's terms.
