# hako-crawler: generalise + PDF export

Two phases. Phase 1 turns a single-novel script into a stable, minimal
`hako-crawler`; phase 2 adds PDF export modelled on `monkeyd-crawler`.

## Outcome

One command downloads any ln.hako.vn novel and writes a phone-readable PDF.
ATNVC ("Anh Trai Nhân Vật Chính") becomes the documented example rather than a
hardcoded URL.

## Constraints

- Go, no browser (headless ARM64 host, no fonts installed -> font must be embedded).
- Keep goquery: selector code is far shorter than hand-walking `html.Node`,
  and it already depends on `golang.org/x/net`.
- Public repo under `tiennm99`.

## Non-goals

- Volume-aware PDF sectioning beyond a per-chapter heading.
- Resuming a partial PDF, or any GUI.

## Site findings (verified 2026-08-24, live fetch)

1. Chapter bodies are no longer plain markup. `#chapter-content` holds
   `<div id="chapter-c-protected" data-s="xor_shuffle" data-k="<16 ascii>"
   data-c="<json array>">`, decoded in-browser by `/scripts/app.js`.
   **The pre-existing `p[id=digits]` extractor therefore returns nothing** —
   the old crawler is already broken against the live site.
2. Decode: sort chunks by the decimal in their first 4 chars, strip that
   prefix, base64-decode, XOR each byte with the key's ASCII bytes cycling —
   the key offset restarts per chunk, not across the joined stream.
3. The decoded payload is the original chapter markup and contains only
   `p`/`em`/`strong`/`img`, every `p` carrying a numeric id. So extraction
   after decoding is a plain `p` walk over content that has no site chrome in it.
4. Landing page: title `.series-name a`, tags `a.series-gerne-item` (site's own
   spelling), chapters `div.chapter-name > a` grouped by `.volume-list`
   sections. Document order is already reading order — unlike monkeydd, no
   reversal is needed.
5. Illustration chapters are real chapters with captions but little prose, so
   an empty-ish chapter must not fail the run.

## Phase 1 — generalise and stabilise

| Step | Detail |
|---|---|
| Module | `github.com/tiennm99/hako-crawler` |
| Layout | `cmd/hako-crawler` CLI, `hako` fetch+parse, `export` pipeline, `pdfout` render |
| Client | one rate-limited, retrying fetcher; 429/5xx retried with backoff, 404/403 fail fast |
| Cache | raw pages under `.cache/`, so re-rendering costs no requests |
| Decode | `hako/protected.go`, with the plain-markup path kept as fallback |
| Validation | `go test ./...` on synthetic fixtures; no network in tests |

## Phase 2 — PDF export

| Step | Detail |
|---|---|
| Render | `pdfout`: 90x160mm phone page default, a5/a4 presets, justified body |
| Font | embedded DejaVu Sans (host has no fonts); explicit `-font` wins and is a hard error |
| Pipeline | `export.Export(ctx, Request)` — list, fetch, font, render |
| Keep | opt-in `-txt <dir>` preserves the original per-chapter text output |

## Acceptance criteria

- `hako-crawler -url <atnvc> -limit 3` writes a PDF with real Vietnamese
  diacritics and non-empty chapters.
- `go test ./...` green, offline.
- `go vet ./...` clean.
- Repo and local folder both named `hako-crawler`; module path matches.

## Risk / rollback

Decoding is tied to a site scheme that can rotate; `data-s` is read from the
page and an unknown scheme fails with a clear message rather than emitting
garbage. Rollback is `git revert` — the rename is a `gh repo rename` back.
