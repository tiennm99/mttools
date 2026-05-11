# image-resizer

Lightweight image batch-resizer written in Go — scales images to a percentage of their original size while preserving aspect ratio.

## Build

```bash
go build -o image-resizer
```

## Usage

Edit the `main()` function in `main.go` to set your source folder and scale percentage, then run:

```bash
go run main.go
```

Example — resize all images in a folder to 50%:

```go
folderPath   := "/path/to/images"
scalePercent := 50    // 50% of original dimensions
overwrite    := true  // overwrite originals; false saves to output folder
```

Output:
```
Image Resizer - Scaling to 50%
Source folder: /path/to/images
Mode: Overwriting original images
--------------------------------------------------
Resized: banner.png -> banner.png
Resized: icon.jpg -> icon.jpg
--------------------------------------------------
Done!
```

## Supported formats

| Input | Output |
|-------|--------|
| JPEG / JPG | JPEG (quality 95) |
| PNG | PNG |
| GIF | JPEG (quality 95) |
| BMP | BMP |
| TIFF / TIF | TIFF |

Aspect ratio is always preserved — the scale percentage applies uniformly to both width and height.

## Related

- [webp-converter](https://github.com/tiennm99/webp-converter) — converts WebP images to PNG/JPG (same domain, different operation).

## License

Apache-2.0 — see [LICENSE](LICENSE).
