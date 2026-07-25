# ghibli-gallery-crawler

Image downloader in Go (originally written in Python and JavaScript in 2020-08-31).

## Install

```bash
go build -o ghibli-gallery-crawler .
```

Or run without building: `go run . <command>`.

## Usage

Two commands, one per behavior carried over from the original scripts.

### `gallery` — numbered ghibli.jp galleries

Downloads `<film>001.jpg` … `<film>050.jpg` for each film, into one directory per film:

```bash
# All 8 default films into ./marnie, ./kaguyahime, ...
ghibli-gallery-crawler gallery

# Pick films, image count, and output location
ghibli-gallery-crawler gallery -films ponyo,chihiro -count 50 -out ./images
```

| Flag | Default | Meaning |
|---|---|---|
| `-films` | the 8 films below | comma-separated film slugs |
| `-count` | `50` | highest image number tried per film |
| `-out` | `.` | where the per-film directories are created |
| `-base` | `https://www.ghibli.jp/gallery/` | base gallery URL |
| `-concurrency` | `8` | parallel downloads |
| `-timeout` | `30s` | per-request timeout |
| `-user-agent` | tool identifier | `User-Agent` header to send |

Default films: `marnie`, `kaguyahime`, `kazetachinu`, `kokurikozaka`, `karigurashi`, `ponyo`, `ged`, `chihiro`.

Image numbers are a fixed range, so films with fewer than `-count` images are normal — those are reported as `missing`, not as failures.

### `scrape` — every image on a page

Parses a page's HTML, collects every `<img src>`, resolves relative URLs, and downloads them:

```bash
# Saves into ./en.wikipedia.org (the URL's host)
ghibli-gallery-crawler scrape https://en.wikipedia.org/wiki/Studio_Ghibli

# Explicit output directory
ghibli-gallery-crawler scrape https://example.com -path ./images
```

Flags: `-path` (default: the URL's host), plus `-concurrency`, `-timeout`, and `-user-agent` as above.

Query strings are stripped from image URLs so that names like `/hsts-pixel.gif?c=3.2.5` produce clean file names. `data:` URIs and duplicates are skipped.

## Output behavior

Images are named after the last path segment of their URL. Existing files with the same name are overwritten. Missing directories are created. Non-200 responses are never written to disk, so error pages cannot land in your output as `.jpg` files.

Note that `ghibli.jp` serves gallery images to any client but returns `403` for its HTML pages unless the request looks like a browser, so `gallery` works there while `scrape` does not.

## License

Apache-2.0
