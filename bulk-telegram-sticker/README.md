# Bulk Telegram Sticker Pack Creator (Go)

Small, Go-based tool to prepare images and upload a Telegram sticker pack.

## The Original Python Version

*In 2025, I rewrote this project using Go. The original Python version of this project can be found at the `feature/python` branch.*

**Note: This Go version isn't runnable, check Python version if you really want to use. Telegram Stickers now have app with convenient features, so maybe you don't need these scripts anymore.**

## Prerequisites
- Go 1.24+ installed
- Create a `.env` in the project root with `API_ID`, `API_HASH`, and `PHONE` (see `.env.example`).

## Quick steps
1. Ensure module deps are present:
   ```pwsh
   go mod tidy
   ```
2. Prepare images (reads `input/`, writes `output/`):
   ```pwsh
   go run prepare.go
   ```
   This produces WebP 512×512 files and `output/sticker_config.json`.
3. Edit `output/sticker_config.json` to set `title`, `short_name` (must be unique), and emojis.
4. Upload the pack via the Go program:
   ```pwsh
   go run main.go
   ```

## Notes
- `short_name` must be unique across Telegram.
- If `@Stickers` asks for additional steps, follow its prompts to finish manually.
