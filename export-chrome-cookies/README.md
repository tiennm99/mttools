# export-chrome-cookies

Python script to export Google Chrome cookies on Windows.

## Quick start

```bash
pip install -r requirements.txt
python main.py
```

> **Note:** Run on Windows with Chrome fully closed. Chrome locks its SQLite cookie database while running.

## Output format

Cookies are written to `cookies.json` in the current directory. Each entry follows the WebExtension cookie schema:

```json
{
  "domain": ".example.com",
  "name": "session_id",
  "value": "abc123",
  "path": "/",
  "secure": true,
  "httpOnly": false,
  "sameSite": "lax",
  "expirationDate": 1735689600,
  "session": false
}
```

## Chrome v10+ encryption

Chrome 80+ on Windows encrypts cookie values using AES-256-GCM with a key stored in `Local State` (protected by DPAPI). The script:
1. Reads the base64-encoded encrypted key from `AppData\Local\Google\Chrome\User Data\Local State`.
2. Decrypts it with `win32crypt.CryptUnprotectData`.
3. Uses the key to decrypt each cookie value from the `Cookies` SQLite database.

You must run the script **as the same Windows user** who owns the Chrome profile — DPAPI keys are user-scoped.

## License

Apache-2.0 — see [LICENSE](LICENSE).
