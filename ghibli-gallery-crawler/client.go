package main

import (
	"net/http"
	"time"
)

// defaultUserAgent identifies this tool to servers. Some sites reject requests
// that do not send a User-Agent header at all, and an honest identifier lets
// operators see who is fetching. Override it with -user-agent.
const defaultUserAgent = "ghibli-gallery-crawler/1.0 (+https://github.com/tiennm99/mttools/tree/main/ghibli-gallery-crawler)"

// userAgentTransport stamps every request with a User-Agent header.
type userAgentTransport struct {
	userAgent string
	base      http.RoundTripper
}

func (t *userAgentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone before mutating: RoundTrippers must not modify the caller's request.
	clone := req.Clone(req.Context())
	clone.Header.Set("User-Agent", t.userAgent)
	return t.base.RoundTrip(clone)
}

// newClient builds an HTTP client that applies the given timeout to each
// request and identifies itself with userAgent.
func newClient(timeout time.Duration, userAgent string) *http.Client {
	return &http.Client{
		Timeout: timeout,
		Transport: &userAgentTransport{
			userAgent: userAgent,
			base:      http.DefaultTransport,
		},
	}
}
