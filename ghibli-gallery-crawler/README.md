# download-images

Simple image downloader (from 2020-08-31). Both Python and JavaScript variants included.

## Usage

```bash
# Python
python3 download_images.py

# JavaScript
npm install
node download_images.js
```

Edit the script to set the source URLs and target directory before running.

## Input format

Both variants read a hardcoded list of image URLs defined at the top of each script. Edit the URL array in `download_images.py` or `download_images.js` to point at your images and set the output directory path.

## Output behavior

Images are saved into the configured output directory, named by their original filename from the URL. Existing files are overwritten. The Python variant uses `requests`; the JS variant uses `axios` + `fs.createWriteStream`.

## License

Apache-2.0
