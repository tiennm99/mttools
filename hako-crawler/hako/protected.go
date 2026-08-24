package hako

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"unicode/utf8"
)

// The site no longer ships chapter text as markup. The body arrives as a
// container carrying an encoded payload, which its own JavaScript turns back
// into the original HTML in the browser:
//
//	<div id="chapter-c-protected" data-s="xor_shuffle"
//	     data-k="6b9dd83fad5a3169" data-c="[&quot;0001BVsF…&quot;, …]">
//
// data-c is a JSON array of chunks in shuffled order, each prefixed with its
// own 4-digit position. Reading DOM text alone yields an empty chapter, which
// is why the plain-paragraph extractor this project started with stopped
// returning anything.
const (
	protectedID = "chapter-c-protected"

	// chunkIndexLen is the width of the decimal position prefix on each chunk.
	chunkIndexLen = 4

	schemeXORShuffle    = "xor_shuffle"
	schemeBase64        = "base64"
	schemeBase64Reverse = "base64_reverse"
)

// payload is the encoded chapter body read off the container's attributes.
type payload struct {
	Scheme string // data-s
	Key    string // data-k
	Chunks []string
}

// parsePayload reads the attribute trio into a payload. The document parser has
// already turned &quot; back into a quote, so data-c is plain JSON here.
func parsePayload(scheme, key, chunksJSON string) (*payload, error) {
	var chunks []string
	if err := json.Unmarshal([]byte(chunksJSON), &chunks); err != nil {
		return nil, fmt.Errorf("parse protected chunk list: %w", err)
	}
	if scheme == "" {
		scheme = schemeBase64
	}
	return &payload{Scheme: scheme, Key: key, Chunks: chunks}, nil
}

// decode reassembles the chapter HTML.
//
// An unrecognised scheme is an error rather than a best-effort guess: the
// alternative is emitting a chapter of mojibake that looks like a successful
// export. When the site rotates its scheme this is the message that says so.
func (p *payload) decode() (string, error) {
	if len(p.Chunks) == 0 {
		return "", fmt.Errorf("protected body carries no chunks")
	}
	switch p.Scheme {
	case schemeXORShuffle, schemeBase64, schemeBase64Reverse:
	default:
		return "", fmt.Errorf("unknown content scheme %q (the site's encoding changed)", p.Scheme)
	}
	if p.Scheme == schemeXORShuffle && p.Key == "" {
		return "", fmt.Errorf("scheme %s needs a key but the page carried none", p.Scheme)
	}

	ordered, err := sortChunks(p.Chunks)
	if err != nil {
		return "", err
	}

	var out []byte
	for i, chunk := range ordered {
		part, err := p.decodeChunk(chunk)
		if err != nil {
			return "", fmt.Errorf("chunk %d of %d: %w", i+1, len(ordered), err)
		}
		out = append(out, part...)
	}

	// Each chunk is decoded on its own, so a multi-byte rune split across a
	// chunk boundary would only show up once they are joined.
	if !utf8.Valid(out) {
		return "", fmt.Errorf("decoded body is not valid UTF-8 (wrong key or changed scheme)")
	}
	return string(out), nil
}

// sortChunks puts the shuffled chunks back in order and strips the position
// prefix that carried the ordering.
func sortChunks(chunks []string) ([]string, error) {
	type indexed struct {
		pos  int
		data string
	}

	items := make([]indexed, 0, len(chunks))
	for _, chunk := range chunks {
		if len(chunk) < chunkIndexLen {
			return nil, fmt.Errorf("chunk shorter than its %d-char position prefix", chunkIndexLen)
		}
		pos, err := strconv.Atoi(chunk[:chunkIndexLen])
		if err != nil {
			return nil, fmt.Errorf("chunk position prefix %q is not a number", chunk[:chunkIndexLen])
		}
		items = append(items, indexed{pos: pos, data: chunk[chunkIndexLen:]})
	}
	sort.SliceStable(items, func(i, j int) bool { return items[i].pos < items[j].pos })

	ordered := make([]string, len(items))
	for i, item := range items {
		ordered[i] = item.data
	}
	return ordered, nil
}

// decodeChunk turns one chunk back into its slice of the chapter HTML.
func (p *payload) decodeChunk(chunk string) ([]byte, error) {
	if p.Scheme == schemeBase64Reverse {
		chunk = reverseASCII(chunk)
	}
	data, err := base64.StdEncoding.DecodeString(chunk)
	if err != nil {
		return nil, fmt.Errorf("base64: %w", err)
	}
	if p.Scheme == schemeXORShuffle {
		// The key cycles from the start of every chunk, not across the joined
		// stream — the site decodes each chunk independently.
		for i := range data {
			data[i] ^= p.Key[i%len(p.Key)]
		}
	}
	return data, nil
}

// reverseASCII reverses a string by byte. base64 is ASCII-only, so this cannot
// split a rune; the site reverses the same way (a JS split/reverse/join).
func reverseASCII(s string) string {
	b := []byte(s)
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}
	return string(b)
}
