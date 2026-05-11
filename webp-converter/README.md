# webp-converter

CLI tool written in Go to convert WebP images to other formats (PNG, JPG, etc.).

## Quick start

```bash
go install github.com/tiennm99/webp-converter@latest
webp-converter -input image.webp -output image.png
```

Or build from source:

```bash
go build -o webp-converter .
./webp-converter -input image.webp -output image.png
```

Docker:

```bash
docker build -t webp-converter .
docker run --rm -v $(pwd):/data webp-converter -input /data/image.webp -output /data/image.png
```

## License

Apache-2.0 — see [LICENSE](LICENSE).
