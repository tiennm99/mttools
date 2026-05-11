# miti-scraper

Configurable single-site web scraper in Go. Crawls one root URL, follows internal links that match whitelist regexes, strips HTML to plain text, and writes one `.txt` per page under `data/`.

Built on [`gocolly/colly`](https://github.com/gocolly/colly).

## Config — `config.yaml`

```yaml
root_url: "https://example.com/"

# Only URLs matching at least one regex are crawled and saved
whitelist:
  - "^https?://([^/]*\\.)?example\\.com(/[^?]*)?$"

# Newline-delimited list of already-processed URLs. Auto-resumes across runs.
data_file: "processed_urls.txt"

# Politeness — seconds between requests
delay_seconds: 1
```

## Run

```bash
go run .
```

## Output

- `data/<urlhost>_<urlpath>.txt` — one file per scraped page, HTML stripped to whitespace-collapsed text (drops `<script>`, `<style>`, `<noscript>`, `<head>`)
- `processed_urls.txt` — running ledger of visited URLs; re-running skips them

## Error handling

Logs (does not abort) on `301/302/303/307/308 REDIRECT`, `403 BLOCKED`, `404 NOT_FOUND`, `429 RATE_LIMITED`, network errors.

## License

Apache-2.0 — see [LICENSE](LICENSE).
