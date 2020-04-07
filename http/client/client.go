package client

import (
	"net"
	"net/http"
	"time"
)

// DefaultHTTPClient - default http client
func DefaultHTTPClient(maxWorkers int) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   90 * time.Second,
				KeepAlive: 60 * time.Second,
			}).DialContext,
			MaxIdleConns:          128,
			MaxIdleConnsPerHost:   maxWorkers + 1,   // one more than needed
			IdleConnTimeout:       90 * time.Second, // from DefaultTransport
			TLSHandshakeTimeout:   10 * time.Second, // from DefaultTransport
			ExpectContinueTimeout: 1 * time.Second,  // from DefaultTransport
		},
	}
}
