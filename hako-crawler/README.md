# atnvc-crawler

Crawl data of "Anh trai nhân vật chính", a novel from https://ln.hako.vn/sang-tac/8476-kiep-nay-la-anh-trai-cua-nhan-vat-chinh

## The Original Python Version

*In 2025, I rewrote this project using Go. The original Python version of this project can be found at the `feature/python` branch.*

## Quick Start

1. Install Go: https://go.dev/dl/
2. `go mod init atnvc-crawler` (if needed)
3. `go get github.com/PuerkitoBio/goquery`
4. `go run main.go`

Data saved in `./data/` as `<chapter-title>.txt`.

## Rate Limits (HTTP 429)
Skip crawled chapters and retry later. Uncomment line 104, replace 100 with number of chapters you want to skip.
```go
chapters = chapters.Slice(50, chapters.Length())
```

## Customize
This version should work with other novels on Hako with some modifications. Feel free to try it!
