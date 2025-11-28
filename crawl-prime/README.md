# crawl-prime

Crawl all prime numbers available on http://compoasso.free.fr/primelistweb/page/prime/liste_online_en.php

*In 2025, I rewrite this project using Go. The Python version of this project can be found at [feature/python](https://github.com/tiennm99/crawl-prime/tree/feature/python) branch.*

## Requirements

- Go 1.18 or higher
- Internet connection to access the prime list website
- Write permissions in the current directory (for `primes.txt`)

## Installation

1. Clone this repository:
   ```bash
   git clone https://github.com/tiennm99/crawl-prime.git
   cd crawl-prime
   ```

2. Install Go (if not already installed):
   ```bash
   # Download and install Go from https://golang.org/dl/
   # Or use your package manager:
   # Ubuntu/Debian: sudo apt install golang
   # macOS: brew install go
   # Windows: Download from official website
   ```

3. Install dependencies:
   ```bash
   go mod tidy
   ```

4. Run the script:
   ```bash
   go run main.go
   ```

## Usage

The program will automatically:
- Start crawling from prime number 1
- Process pages in batches of 600 primes each
- Write all found primes to `primes.txt`
- Display progress information during execution
- Stop when it reaches 1,000,000 primes or encounters an error

## Output

- `primes.txt` - Contains all crawled prime numbers, one per line

## License

This project is for educational purposes. Please respect the terms of service of the prime list website.
