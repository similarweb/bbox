package testutils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
)

// SetupMockServer initializes a mock server with configurable endpoints.
func SetupMockServer(config MockServerConfig) (*httptest.Server, *url.URL) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, endpoint := range config.Endpoints {
			if r.URL.Path == endpoint.Path && r.Method == endpoint.Method {
				w.WriteHeader(endpoint.StatusCode)
				if endpoint.Response != nil {
					json.NewEncoder(w).Encode(endpoint.Response)
				}
				return
			}
		}
		// Default to 404 Not Found if no endpoint matches
		http.NotFound(w, r)
	}))

	baseURL, _ := url.Parse(server.URL)
	return server, baseURL
}
