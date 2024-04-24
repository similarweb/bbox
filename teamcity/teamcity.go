package teamcity

import (
	"bytes"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"net/url"
	"strings"
)

var (
	_ IArtifactsService = &ArtifactsService{}
	_ IQueueService     = &QueueService{}
	_ IBuildService     = &BuildService{}
)

type Client struct {
	baseURL   *url.URL
	client    *http.Client
	BasicAuth *BasicAuth

	common service
	// Services of Teamcity
	Artifacts IArtifactsService
	Queue     IQueueService
	Build     IBuildService
}

type BasicAuth struct {
	username string
	password string
}

type service struct {
	client *Client
}

// NewTeamCityClient creates a new TeamCity client.
func NewTeamCityClient(baseURL *url.URL, username, password string) *Client {
	newClient := &Client{
		baseURL: baseURL,
		BasicAuth: &BasicAuth{
			username: username,
			password: password,
		},
		client: &http.Client{},
	}

	newClient.initializeServices()

	return newClient
}

//
//// NewTeamCityClient creates a new TeamCity client with dependencies injected.
//func NewTeamCityClient(baseURL *url.URL, username, password string, httpClient *http.Client, artifacts IArtifactsService, queue IQueueService, build IBuildService) *Client {
//	if httpClient == nil {
//		httpClient = &http.Client{} // Default to http.Client if none provided
//	}
//	return &Client{
//		baseURL:   baseURL,
//		client:    httpClient,
//		BasicAuth: &BasicAuth{username: username, password: password},
//		Artifacts: artifacts,
//		Queue:     queue,
//		Build:     build,
//	}
//}

func (c *Client) initializeServices() {
	c.common.client = c
	c.Artifacts = &ArtifactsService{client: c}
	c.Queue = &QueueService{client: c}
	c.Build = &BuildService{client: c}
}

// RequestOption represents an option that can modify an http.Request.
type RequestOption func(req *http.Request)

// NewRequestWrapper creates an API request with basic Auth with Username and Password.
// This Function injects the Accept and Content-Type headers.
// A relative URL can be provided in urlStr, in which case it is resolved relative to the BaseURL of the Client.
func (c *Client) NewRequestWrapper(method, urlStr string, body interface{}, opts ...RequestOption) (*http.Request, error) {
	if !strings.HasSuffix(c.baseURL.Path, "/") {
		// add trailing slash to baseURL
		c.baseURL.Path += "/"
		// return nil, fmt.Errorf("BaseURL must have a trailing slash, but %v does not", c.baseURL)
	}

	u, err := c.baseURL.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse url %s: %w", urlStr, err)
	}

	log.WithFields(log.Fields{
		"method": method,
		"url":    u.String(),
	}).Debug("creating new request")

	var buf io.ReadWriter
	if body != nil {
		buf = &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(false)

		err := enc.Encode(body)
		if err != nil {
			return nil, fmt.Errorf("failed to encode body: %w", err)
		}
	}

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create new request: %w", err)
	}

	req.SetBasicAuth(c.BasicAuth.username, c.BasicAuth.password)

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	req.Header.Set("Accept", "application/json")

	for _, opt := range opts {
		opt(req)
	}

	return req, nil
}
