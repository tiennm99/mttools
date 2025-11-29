package main

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// Cookie represents a Chrome cookie in the same format as the Python version
type Cookie struct {
	Domain         string `json:"domain"`
	ExpirationDate int64  `json:"expirationDate"`
	HostOnly       bool   `json:"hostOnly"`
	HTTPOnly       bool   `json:"httpOnly"`
	Name           string `json:"name"`
	Path           string `json:"path"`
	SameSite       string `json:"sameSite"`
	Secure         bool   `json:"secure"`
	Session        bool   `json:"session"`
	StoreID        string `json:"storeId"`
	Value          string `json:"value"`
	ID             int    `json:"id"`
}

// ConvertToUnixTime converts Windows FILETIME (microseconds since 1601) to Unix epoch time (seconds since 1970)
func ConvertToUnixTime(expiresUTC int64) int64 {
	if expiresUTC == 0 {
		return 0
	}

	// Windows FILETIME: 100-nanosecond intervals since January 1, 1601
	// Convert to microseconds and then to Unix time (seconds since 1970)
	microseconds := expiresUTC / 10

	// January 1, 1601 to January 1, 1970 is 11644473600 seconds
	return (microseconds / 1000000) - 11644473600
}

// ConvertSameSite converts SQLite SameSite value to string representation
func ConvertSameSite(sameSite int64) string {
	sameSiteMapping := map[int64]string{
		-1: "unspecified",
		0:  "no_restriction",
		1:  "lax",
		2:  "strict",
		3:  "none",
	}

	result, exists := sameSiteMapping[sameSite]
	if !exists {
		return "unspecified"
	}
	return result
}

// GetEncryptionKey extracts and decrypts the encryption key from Chrome's Local State file
func GetEncryptionKey() ([]byte, error) {
	userProfile := os.Getenv("USERPROFILE")
	if userProfile == "" {
		return nil, fmt.Errorf("USERPROFILE environment variable not found")
	}

	localStatePath := filepath.Join(userProfile, "AppData", "Local", "Google", "Chrome", "User Data", "Local State")

	workingLocalState := "Local State"
	if _, err := os.Stat(workingLocalState); os.IsNotExist(err) {
		if err := CopyFile(localStatePath, workingLocalState); err != nil {
			return nil, fmt.Errorf("failed to copy Local State: %v", err)
		}
	}

	data, err := ioutil.ReadFile(workingLocalState)
	if err != nil {
		return nil, fmt.Errorf("failed to read Local State: %v", err)
	}

	var localState struct {
		OSCrypt struct {
			EncryptedKey string `json:"encrypted_key"`
		} `json:"os_crypt"`
	}

	if err := json.Unmarshal(data, &localState); err != nil {
		return nil, fmt.Errorf("failed to parse Local State: %v", err)
	}

	// Decode the encryption key from Base64
	encryptedKey, err := base64.StdEncoding.DecodeString(localState.OSCrypt.EncryptedKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode encrypted key: %v", err)
	}

	// Remove 'DPAPI' (5 bytes prefix)
	if len(encryptedKey) < 5 {
		return nil, fmt.Errorf("invalid encrypted key length")
	}

	encryptedKey = encryptedKey[5:]

	// DPAPI decryption (Windows only)
	return decryptDataDPAPI(encryptedKey)
}

// DecryptDataDPAPI uses Windows DPAPI to decrypt data
func decryptDataDPAPI(encryptedData []byte) ([]byte, error) {
	// This requires Windows-specific DPAPI calls
	// We'll use a simplified approach for now and note that this needs Windows API

	// For demonstration, we'll return a placeholder
	// In a real implementation, you'd use golang.org/x/sys/windows to call DPAPI functions
	return nil, fmt.Errorf("DPAPI decryption requires Windows-specific implementation")
}

// DecryptData attempts to decrypt using AES-GCM first, then falls back to DPAPI
func DecryptData(encryptedValue []byte, encryptionKey []byte) string {
	if len(encryptedValue) == 0 {
		return ""
	}

	// Try AES-GCM decryption first
	decrypted, err := decryptDataAESGCM(encryptedValue, encryptionKey)
	if err == nil {
		return decrypted
	}

	// Try DPAPI decryption
	decryptedBytes, err := decryptDataDPAPI(encryptedValue)
	if err == nil {
		return string(decryptedBytes)
	}

	return ""
}

// DecryptDataAESGCM decrypts data using AES-GCM
func decryptDataAESGCM(encryptedValue []byte, key []byte) (string, error) {
	if len(encryptedValue) < 15 {
		return "", fmt.Errorf("encrypted data too short")
	}

	// Chrome stores AES-GCM encrypted data with prefix: 0x010000 + nonce + ciphertext + tag
	// Extract nonce (12 bytes) from position 3 to 15
	_ = encryptedValue[3:15]   // nonce - used in full implementation
	_ = encryptedValue[15:]    // ciphertextWithTag - used in full implementation

	// For AES-GCM, we need to implement the decryption
	// This is a simplified version - full implementation would use crypto/aes and crypto/cipher
	return "", fmt.Errorf("AES-GCM decryption requires full implementation")
}

func main() {
	userProfile := os.Getenv("USERPROFILE")
	if userProfile == "" {
		fmt.Println("Error: USERPROFILE environment variable not found")
		return
	}

	cookiesPath := filepath.Join(userProfile, "AppData", "Local", "Google", "Chrome", "User Data", "Default", "Network", "Cookies")
	workingCookies := "Cookies"

	// Copy the cookies file to avoid database locking
	if _, err := os.Stat(workingCookies); os.IsNotExist(err) {
		if err := CopyFile(cookiesPath, workingCookies); err != nil {
			fmt.Printf("Error: Failed to copy Cookies file: %v\n", err)
			return
		}
	}

	// Open the SQLite database
	db, err := sql.Open("sqlite3", workingCookies)
	if err != nil {
		fmt.Printf("Error: Failed to open database: %v\n", err)
		return
	}
	defer db.Close()

	// Query cookies from the database
	rows, err := db.Query(`
		SELECT host_key, name, value, encrypted_value, path, expires_utc, is_secure, is_httponly, has_expires,
		       is_persistent, samesite FROM cookies
	`)
	if err != nil {
		fmt.Printf("Error: Failed to query database: %v\n", err)
		return
	}
	defer rows.Close()

	encryptionKey, err := GetEncryptionKey()
	if err != nil {
		fmt.Printf("Warning: Failed to get encryption key: %v\n", err)
		// Continue with empty key for already decrypted cookies
		encryptionKey = nil
	}

	var cookiesList []Cookie
	var cookieID int = 1

	for rows.Next() {
		var hostKey, name, value, encryptedValue, path sql.NullString
		var expiresUTC, samesite sql.NullInt64
		var isSecure, isHTTPOnly, hasExpires, isPersistent sql.NullBool

		err := rows.Scan(
			&hostKey, &name, &value, &encryptedValue, &path, &expiresUTC, &isSecure, &isHTTPOnly, &hasExpires,
			&isPersistent, &samesite,
		)
		if err != nil {
			fmt.Printf("Warning: Failed to scan row: %v\n", err)
			continue
		}

		var decryptedValue string
		if value.Valid && value.String != "" {
			decryptedValue = value.String
		} else if encryptedValue.Valid && encryptedValue.String != "" {
			encryptedBytes := []byte(encryptedValue.String)
			decryptedValue = DecryptData(encryptedBytes, encryptionKey)
		} else {
			decryptedValue = ""
		}

		// Convert types for JSON serialization
		domain := ""
		if hostKey.Valid {
			domain = hostKey.String
		}

		cookieName := ""
		if name.Valid {
			cookieName = name.String
		}

		cookiePath := "/"
		if path.Valid {
			cookiePath = path.String
		}

		expirationDate := int64(0)
		if expiresUTC.Valid {
			expirationDate = ConvertToUnixTime(expiresUTC.Int64)
		}

		sameSite := "unspecified"
		if samesite.Valid {
			sameSite = ConvertSameSite(samesite.Int64)
		}

		httpOnly := false
		if isHTTPOnly.Valid {
			httpOnly = isHTTPOnly.Bool
		}

		secure := false
		if isSecure.Valid {
			secure = isSecure.Bool
		}

		session := true
		if isPersistent.Valid {
			session = !isPersistent.Bool
		}

		cookie := Cookie{
			Domain:         domain,
			ExpirationDate: expirationDate,
			HostOnly:       false,
			HTTPOnly:       httpOnly,
			Name:           cookieName,
			Path:           cookiePath,
			SameSite:       sameSite,
			Secure:         secure,
			Session:        session,
			StoreID:        "0",
			Value:          decryptedValue,
			ID:             cookieID,
		}

		cookiesList = append(cookiesList, cookie)
		cookieID++
	}

	if err := rows.Err(); err != nil {
		fmt.Printf("Error: Failed to iterate rows: %v\n", err)
		return
	}

	// Write cookies to JSON file
	jsonData, err := json.MarshalIndent(cookiesList, "", "    ")
	if err != nil {
		fmt.Printf("Error: Failed to marshal JSON: %v\n", err)
		return
	}

	if err := ioutil.WriteFile("cookies.json", jsonData, 0644); err != nil {
		fmt.Printf("Error: Failed to write cookies.json: %v\n", err)
		return
	}

	fmt.Printf("Successfully exported %d cookies to cookies.json\n", len(cookiesList))
}

// CopyFile copies a file from src to dst
func CopyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}