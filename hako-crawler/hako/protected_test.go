package hako

import (
	"encoding/base64"
	"fmt"
	"strings"
	"testing"
)

// encodeXORShuffle builds a payload the way the site does: split the body into
// chunks, XOR each one from the start of the key, base64 it, and prefix its
// position. Chunks are emitted out of order on purpose — the shuffle is the
// part the decoder has to undo.
func encodeXORShuffle(t *testing.T, body, key string, chunkLen int) []string {
	t.Helper()

	var chunks []string
	raw := []byte(body)
	for pos := 0; pos*chunkLen < len(raw); pos++ {
		start := pos * chunkLen
		end := min(start+chunkLen, len(raw))

		part := make([]byte, end-start)
		copy(part, raw[start:end])
		for i := range part {
			part[i] ^= key[i%len(key)]
		}
		chunks = append(chunks, fmt.Sprintf("%04d%s", pos, base64.StdEncoding.EncodeToString(part)))
	}

	// Reverse so the decoder cannot pass by accident on input order alone.
	for i, j := 0, len(chunks)-1; i < j; i, j = i+1, j-1 {
		chunks[i], chunks[j] = chunks[j], chunks[i]
	}
	return chunks
}

func jsonArray(items []string) string {
	quoted := make([]string, len(items))
	for i, s := range items {
		quoted[i] = `"` + s + `"`
	}
	return "[" + strings.Join(quoted, ",") + "]"
}

// The body is deliberately split so multi-byte runes straddle chunk
// boundaries, which is where a decoder that XORs across the joined stream, or
// that decodes each chunk as its own UTF-8 string, falls apart.
const (
	testKey  = "6b9dd83fad5a3169"
	testBody = `<p id="1">Trời hôm nay đẹp lắm.</p><p id="2">Người anh trai mỉm cười.</p>`
)

func TestDecodeXORShuffle(t *testing.T) {
	for _, chunkLen := range []int{7, 16, 31, 4096} {
		load, err := parsePayload(schemeXORShuffle, testKey,
			jsonArray(encodeXORShuffle(t, testBody, testKey, chunkLen)))
		if err != nil {
			t.Fatalf("chunk length %d: parse payload: %v", chunkLen, err)
		}
		got, err := load.decode()
		if err != nil {
			t.Fatalf("chunk length %d: decode: %v", chunkLen, err)
		}
		if got != testBody {
			t.Errorf("chunk length %d: body round-trip failed\n got: %q\nwant: %q", chunkLen, got, testBody)
		}
	}
}

func TestDecodeRejectsWrongKey(t *testing.T) {
	load, err := parsePayload(schemeXORShuffle, "0000000000000000",
		jsonArray(encodeXORShuffle(t, testBody, testKey, 16)))
	if err != nil {
		t.Fatalf("parse payload: %v", err)
	}
	// A wrong key yields bytes that are not valid UTF-8; the decoder must say
	// so rather than hand back a chapter of mojibake.
	if _, err := load.decode(); err == nil {
		t.Fatal("expected an error for a wrong key, got none")
	}
}

func TestDecodeRejectsUnknownScheme(t *testing.T) {
	// When the site rotates its encoding, failing loudly here is what tells the
	// user why the export stopped working.
	load, err := parsePayload("aes_gcm_2027", testKey, `["0000AAAA"]`)
	if err != nil {
		t.Fatalf("parse payload: %v", err)
	}
	if _, err := load.decode(); err == nil {
		t.Fatal("expected an error for an unknown scheme, got none")
	}
}

func TestDecodeRejectsMissingKey(t *testing.T) {
	load, err := parsePayload(schemeXORShuffle, "", `["0000AAAA"]`)
	if err != nil {
		t.Fatalf("parse payload: %v", err)
	}
	if _, err := load.decode(); err == nil {
		t.Fatal("expected an error when the key is absent, got none")
	}
}

func TestDecodeRejectsEmptyPayload(t *testing.T) {
	load, err := parsePayload(schemeXORShuffle, testKey, `[]`)
	if err != nil {
		t.Fatalf("parse payload: %v", err)
	}
	if _, err := load.decode(); err == nil {
		t.Fatal("expected an error for an empty chunk list, got none")
	}
}

// base64 and base64_reverse are the site's simpler schemes; the decoder still
// carries them so a chapter served either way keeps working.
func TestDecodePlainBase64Schemes(t *testing.T) {
	encoded := base64.StdEncoding.EncodeToString([]byte(testBody))

	for _, tc := range []struct {
		scheme string
		chunk  string
	}{
		{schemeBase64, encoded},
		{schemeBase64Reverse, reverseASCII(encoded)},
	} {
		load, err := parsePayload(tc.scheme, "", jsonArray([]string{"0000" + tc.chunk}))
		if err != nil {
			t.Fatalf("%s: parse payload: %v", tc.scheme, err)
		}
		got, err := load.decode()
		if err != nil {
			t.Fatalf("%s: decode: %v", tc.scheme, err)
		}
		if got != testBody {
			t.Errorf("%s: got %q, want %q", tc.scheme, got, testBody)
		}
	}
}

func TestSortChunksRejectsBadPrefix(t *testing.T) {
	if _, err := sortChunks([]string{"abcdAAAA"}); err == nil {
		t.Fatal("expected an error for a non-numeric position prefix, got none")
	}
	if _, err := sortChunks([]string{"01"}); err == nil {
		t.Fatal("expected an error for a chunk shorter than its prefix, got none")
	}
}
