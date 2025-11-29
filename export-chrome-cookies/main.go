package main

import (
	"crypto/aes"
	"crypto/cipher"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
	_ "modernc.org/sqlite"
)

type Cookie struct {
	Domain         string  `json:"domain"`
	ExpirationDate float64 `json:"expirationDate"`
	HostOnly       bool    `json:"hostOnly"`
	HTTPOnly       bool    `json:"httpOnly"`
	Name           string  `json:"name"`
	Path           string  `json:"path"`
	SameSite       string  `json:"sameSite"`
	Secure         bool    `json:"secure"`
	Session        bool    `json:"session"`
	StoreID        string  `json:"storeId"`
	Value          string  `json:"value"`
	ID             int     `json:"id"`
}

type LocalState struct {
	OSCrypt struct {
		EncryptedKey string `json:"encrypted_key"`
	} `json:"os_crypt"`
}

func main() {
	userProfile := os.Getenv("USERPROFILE")
	localStatePath := filepath.Join(userProfile, "AppData", "Local", "Google", "Chrome", "User Data", "Local State")
	cookiesPath := filepath.Join(userProfile, "AppData", "Local", "Google", "Chrome", "User Data", "Default", "Network", "Cookies")

	// Get encryption key
	key, err := getEncryptionKey(localStatePath)
	if err != nil {
		fmt.Printf("Error getting encryption key: %v\n", err)
		return
	}
	fmt.Printf("Encryption key length: %d bytes\n", len(key))
	fmt.Printf("Key first bytes: %v\n", key[:min(16, len(key))])

	// Extract cookies
	cookies, err := extractCookies(cookiesPath, key)
	if err != nil {
		fmt.Printf("Error extracting cookies: %v\n", err)
		return
	}

	// Save to JSON
	jsonData, err := json.MarshalIndent(cookies, "", "    ")
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		return
	}

	err = os.WriteFile("cookies.json", jsonData, 0644)
	if err != nil {
		fmt.Printf("Error writing JSON file: %v\n", err)
		return
	}

	fmt.Printf("Successfully extracted %d cookies to cookies.json\n", len(cookies))
}

func getEncryptionKey(localStatePath string) ([]byte, error) {
	// Always copy a fresh Local State file
	workingFile := "Local State"
	// Remove old copy if exists
	os.Remove(workingFile)

	if err := copyFile(localStatePath, workingFile); err != nil {
		return nil, fmt.Errorf("failed to copy Local State: %v", err)
	}

	// Read and parse Local State
	data, err := os.ReadFile(workingFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read Local State: %v", err)
	}

	var localState LocalState
	if err := json.Unmarshal(data, &localState); err != nil {
		return nil, fmt.Errorf("failed to parse Local State: %v", err)
	}

	// Decode base64 key
	encryptedKey, err := base64.StdEncoding.DecodeString(localState.OSCrypt.EncryptedKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode key: %v", err)
	}

	fmt.Printf("Encrypted key length: %d bytes\n", len(encryptedKey))
	fmt.Printf("Encrypted key prefix: %s\n", string(encryptedKey[:5]))

	// Remove "DPAPI" prefix (first 5 bytes)
	if len(encryptedKey) < 5 || string(encryptedKey[:5]) != "DPAPI" {
		return nil, fmt.Errorf("encrypted key doesn't have DPAPI prefix")
	}
	encryptedKey = encryptedKey[5:]

	// Decrypt using DPAPI
	key, err := dpAPIDecrypt(encryptedKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt key: %v", err)
	}

	return key, nil
}

func dpAPIDecrypt(data []byte) ([]byte, error) {
	dataBlob := windows.DataBlob{
		Size: uint32(len(data)),
		Data: &data[0],
	}
	var outBlob windows.DataBlob

	err := windows.CryptUnprotectData(&dataBlob, nil, nil, 0, nil, 0, &outBlob)
	if err != nil {
		return nil, fmt.Errorf("DPAPI decrypt failed: %v", err)
	}

	decrypted := make([]byte, outBlob.Size)
	copy(decrypted, unsafe.Slice(outBlob.Data, outBlob.Size))
	windows.LocalFree(windows.Handle(unsafe.Pointer(outBlob.Data)))

	return decrypted, nil
}

func extractCookies(cookiesPath string, key []byte) ([]Cookie, error) {
	// Always copy a fresh cookies database
	workingDB := "Cookies"
	// Remove old copy if exists
	os.Remove(workingDB)

	if err := copyFile(cookiesPath, workingDB); err != nil {
		return nil, fmt.Errorf("failed to copy Cookies database: %v", err)
	}

	// Open database
	db, err := sql.Open("sqlite", workingDB)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}
	defer db.Close()

	// Query cookies
	query := `SELECT host_key, name, value, encrypted_value, path, expires_utc, is_secure,
	                 is_httponly, has_expires, is_persistent, samesite FROM cookies`
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query cookies: %v", err)
	}
	defer rows.Close()

	var cookies []Cookie
	id := 1

	for rows.Next() {
		var hostKey, name, value, path string
		var encryptedValue []byte
		var expiresUTC, isSecure, isHTTPOnly, hasExpires, isPersistent, sameSite int64

		err := rows.Scan(&hostKey, &name, &value, &encryptedValue, &path, &expiresUTC,
			&isSecure, &isHTTPOnly, &hasExpires, &isPersistent, &sameSite)
		if err != nil {
			continue
		}

		// Debug output for first few cookies
		if id <= 3 {
			fmt.Printf("Cookie #%d: %s\n", id, name)
			fmt.Printf("  Value field: '%s' (len=%d)\n", value, len(value))
			fmt.Printf("  Encrypted field: len=%d\n", len(encryptedValue))
			if len(encryptedValue) > 0 {
				fmt.Printf("  First bytes: %v\n", encryptedValue[:min(10, len(encryptedValue))])
			}
		}

		// Decrypt value if needed
		decryptedValue := value
		if value == "" && len(encryptedValue) > 0 {
			decryptedValue = decryptData(encryptedValue, key)
			if id <= 3 {
				fmt.Printf("  Decrypted: '%s'\n", decryptedValue)
			}
		}

		cookie := Cookie{
			Domain:         hostKey,
			ExpirationDate: convertToUnixTime(expiresUTC),
			HostOnly:       false,
			HTTPOnly:       isHTTPOnly == 1,
			Name:           name,
			Path:           path,
			SameSite:       convertSameSite(int(sameSite)),
			Secure:         isSecure == 1,
			Session:        isPersistent == 0,
			StoreID:        "0",
			Value:          decryptedValue,
			ID:             id,
		}

		cookies = append(cookies, cookie)
		id++
	}

	return cookies, nil
}

func decryptData(data []byte, key []byte) string {
	// Check if data starts with "v10", "v11", or "v20" prefix (Chrome's encryption marker)
	if len(data) > 3 && data[0] == 'v' {
		version := string(data[0:3])

		// v10, v11, v20 all use AES-GCM with 12-byte nonce
		if version == "v10" || version == "v11" || version == "v20" {
			if len(data) < 15 {
				return ""
			}

			// Nonce is bytes 3-15 (12 bytes)
			nonce := data[3:15]
			encryptedData := data[15:]

			// Create AES-GCM cipher
			block, err := aes.NewCipher(key)
			if err != nil {
				fmt.Printf("  AES cipher error: %v\n", err)
				return ""
			}

			aesgcm, err := cipher.NewGCM(block)
			if err != nil {
				fmt.Printf("  GCM error: %v\n", err)
				return ""
			}

			// Decrypt
			decrypted, err := aesgcm.Open(nil, nonce, encryptedData, nil)
			if err != nil {
				fmt.Printf("  Decrypt error: %v\n", err)
				return ""
			}

			return string(decrypted)
		}
	}

	// Try DPAPI fallback for older encryption
	if len(data) > 0 {
		if decrypted, err := dpAPIDecrypt(data); err == nil {
			return string(decrypted)
		}
	}

	return ""
}

func convertToUnixTime(expiresUTC int64) float64 {
	if expiresUTC == 0 {
		return 0
	}

	// Chrome's timestamp is microseconds since January 1, 1601
	// Convert to Unix timestamp (seconds since January 1, 1970)
	chromeEpoch := time.Date(1601, 1, 1, 0, 0, 0, 0, time.UTC)
	unixEpoch := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)

	expiresTime := chromeEpoch.Add(time.Duration(expiresUTC) * time.Microsecond)
	unixTime := expiresTime.Sub(unixEpoch).Seconds()

	return unixTime
}

func convertSameSite(sameSite int) string {
	sameSiteMap := map[int]string{
		-1: "unspecified",
		0:  "no_restriction",
		1:  "lax",
		2:  "strict",
		3:  "none",
	}

	if val, ok := sameSiteMap[sameSite]; ok {
		return val
	}
	return "unspecified"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
