# hako-crawler

Downloads every chapter of an [ln.hako.vn](https://ln.hako.vn) novel and exports
it as a single PDF sized for reading on a phone.

## Install

Requires Go 1.25+.

```sh
go build ./cmd/hako-crawler
```

## Usage

Pass the novel's landing page URL. "Anh Trai Nhân Vật Chính" — the novel this
project was originally written for — works as the example:

```sh
./hako-crawler -url https://ln.hako.vn/sang-tac/8476-kiep-nay-la-anh-trai-cua-nhan-vat-chinh
```

The PDF is named after the novel unless you pass `-out`.

```sh
# Try the layout on 3 chapters before fetching all 147
./hako-crawler -url <novel URL> -limit 3

# Bigger type on an A5 page for a tablet
./hako-crawler -url <novel URL> -page a5 -font-size 12 -out truyen.pdf

# Also keep one plain-text file per chapter
./hako-crawler -url <novel URL> -txt ./data
```

### Flags

| Flag | Default | Purpose |
| --- | --- | --- |
| `-url` | *required* | Novel landing page URL |
| `-out` | novel name | Output PDF path |
| `-txt` | off | Also write one plain-text file per chapter into this directory |
| `-page` | `phone` | Page size: `phone`, `a5`, `a4` |
| `-font-size` | `10` | Body font size in points |
| `-line-spacing` | `1.55` | Line height as a multiple of font size |
| `-margin` | `6` | Page margin in mm |
| `-font` | auto | Path to a `.ttf`; defaults to a system font, else the bundled one |
| `-workers` | `4` | Concurrent chapter fetches |
| `-delay` | `500ms` | Minimum delay between requests |
| `-retries` | `3` | Retries per request |
| `-limit` | `0` | Fetch only the first N chapters (0 = all) |
| `-cache` | `.cache` | Cache directory for raw pages (empty to disable) |

## Why the default page is 90×160 mm

Phone readability is governed by page *shape* more than by font size. A PDF
viewer scales a whole page to fit the screen, so a large font on an A4 page
still ends up small: the page is about three times wider than a phone screen
and gets shrunk to match. The default page is cut to a 9:16 ratio so it fills
the screen at 100% zoom, where the default 10 pt renders at a comfortable size.
Use `-page a5` or `-page a4` for a tablet or for printing.

## How it works

1. Fetch the landing page: title, genres, and the chapter list, read from the
   `.volume-list` sections. The site lists volumes oldest-first and chapters
   ascending within a volume, so document order is already reading order.
2. Fetch each chapter concurrently and decode its body.
3. Render one PDF, each chapter starting on a new page, with the volume name
   printed above the heading whenever it changes.

### The chapter body is not markup

The site no longer ships chapter text as HTML. `#chapter-content` holds an
encoded payload that its own JavaScript expands in the browser:

```html
<div id="chapter-c-protected" data-s="xor_shuffle"
     data-k="6b9dd83fad5a3169" data-c="[&quot;0001BVsFVFQKD1YNF0Ra…&quot;, …]">
```

`data-c` is a JSON array of chunks in shuffled order. Each chunk is prefixed
with its own 4-digit position; strip that, base64-decode the rest, and XOR every
byte against the ASCII bytes of `data-k`, cycling. The key restarts at the
beginning of *every chunk* rather than running across the joined stream, so
XOR-ing the concatenated payload in one pass produces garbage. Sorting by the
position prefix then joining yields the original chapter HTML, which contains
only `p`, `em`, `strong` and `img` — no site furniture — so extracting the text
afterwards is a plain paragraph walk.

Reading DOM text alone yields an *empty* chapter, which is why the
`p[id=<digits>]` extractor this project started with stopped returning anything.

`data-s` names the scheme and is read from the page rather than assumed. An
unrecognised value is a hard error: the alternative is a book full of mojibake
that looks like a successful export. If the site rotates its encoding, that
error message is what says so.

### Illustration chapters

Hako novels carry illustration chapters that hold pictures and captions rather
than prose. They are real chapters, so a near-empty one is exported with its
heading instead of failing the run.

## Politeness and caching

Requests are spaced by `-delay` globally, so raising `-workers` does not raise
the request rate — which is what keeps a full-novel crawl off the HTTP 429 that
the original single-threaded script kept hitting. A 429 or 5xx is retried with
exponential backoff; a 404 or 403 fails immediately rather than spending the
retry budget on something that will not change.

Raw pages are cached under `.cache/`, so re-exporting with different font or
page settings costs no requests. Delete the directory to refetch.

## Tests

```sh
go test ./...
```

Tests run against synthetic fixtures — including a payload encoded the same way
the site encodes one, with chunks shuffled and multi-byte characters straddling
chunk boundaries. No network access required.

## Layout

```
cmd/hako-crawler/    CLI
export/              URL -> PDF in one call; shared by the CLI and importers
hako/                fetching, page parsing, body decoding, crawl orchestration
pdfout/              PDF rendering, font resolution, bundled fallback font
```

## Use as a library

The packages are importable, so another Go program can produce the same PDF
without shelling out to the binary. `export.Export` is the whole pipeline:

```go
result, err := export.Export(ctx, export.Request{
	NovelURL: "https://ln.hako.vn/sang-tac/8476-kiep-nay-la-anh-trai-cua-nhan-vat-chinh",
	OutDir:   tmpDir,
})
```

Only `NovelURL` is required; each zero-valued field falls back to the same
default as the matching CLI flag. Because zero means "unset", ask for *no* cache
or *no* request delay with the `NoCache` and `NoDelay` fields rather than by
zeroing `CacheDir` or `Delay`. Pass a `Log` function to receive the progress
messages the CLI prints to stderr.

## Fonts

The PDF embeds a TrueType font, and Vietnamese needs one covering the Latin
Extended Additional block — a basic-Latin font silently drops the diacritics.
The font is resolved in this order:

1. the path given to `-font` / `Request.FontFile`, which is an error if it
   cannot be read — a named font is not silently substituted;
2. a system font known to cover Vietnamese (see `pdfout.FindFont`);
3. the bundled DejaVu Sans, compiled into the binary.

Step 3 means rendering never depends on the host having fonts installed, which
is what a minimal container or a headless server usually looks like. See
[`pdfout/fonts/NOTICE.md`](pdfout/fonts/NOTICE.md) for the bundled font's
provenance and licensing.

## History

This started as a Python script for a single novel, was rewritten in Go in 2025,
and became a general hako exporter in 2026. The Python version is on the
`feature/python` branch.

## Scope

Downloaded text stays on your machine. Only fetch content you are allowed to
read offline, and respect the site's terms.
