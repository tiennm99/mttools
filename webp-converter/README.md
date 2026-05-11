# webp-converter

Lightweight HTTP microservice written in Go that converts WebP images to PNG on-the-fly — accepts an image URL, returns the converted PNG bytes.

## Quick start

```bash
go build -o webp-converter .
./webp-converter
# Server listens on :8080
```

Docker:

```bash
docker build -t webp-converter .
docker run --rm -p 8080:8080 webp-converter
```

## Usage

POST a JSON body with the source image URL to `/process`:

```bash
curl -X POST http://localhost:8080/process \
  -H "Content-Type: application/json" \
  -d '{"img": "https://example.com/image.webp"}' \
  --output image.png
```

- If the URL points to a **WebP** image, the response is the converted **PNG**.
- If the URL points to any other format, the original image bytes are returned unchanged (pass-through).

## Output formats

| Input | Output |
|-------|--------|
| WebP | PNG |
| Any other format | Original (pass-through) |

## Related

- [image-resizer](https://github.com/tiennm99/image-resizer) — batch-resize images by percentage (same domain, different operation).

## License

Apache-2.0 — see [LICENSE](LICENSE).
