package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/function61/gokit/net/http/ezhttp"
	"github.com/function61/gokit/testing/assert"
)

func TestUrlShortener(t *testing.T) {
	srv := httptest.NewServer(newServerHandlerWithDb(map[string]string{
		"self-test": "https://example.net/redirect-target",
	}))
	defer srv.Close()

	noFollowRedirects := ezhttp.Client(&http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	})

	_, err := ezhttp.Get(context.Background(), srv.URL+"/", noFollowRedirects)
	assert.Equal(t, err.Error(), "404 Not Found; 404 page not found\n")

	_, err = ezhttp.Get(context.Background(), srv.URL+"/go/link-not-found", noFollowRedirects)
	assert.Equal(t, err.Error(), "404 Not Found; 404 page not found\n")

	resp, err := ezhttp.Get(context.Background(), srv.URL+"/go/self-test", noFollowRedirects, ezhttp.TolerateNon2xxResponse)
	assert.Ok(t, err)

	assert.Equal(t, resp.StatusCode, http.StatusFound)
	assert.Equal(t, resp.Header.Get("Location"), "https://example.net/redirect-target")
}
