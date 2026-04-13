// Package env loads .env files on init so all cmd/ scripts get env vars automatically.
// Usage: import _ "github.com/tiennm99/go-util/internal/env"
package env

import (
	"os"

	"github.com/joho/godotenv"
)

func init() {
	// Skip if .env doesn't exist; load if it does.
	if _, err := os.Stat(".env"); err == nil {
		_ = godotenv.Load()
	}
}
