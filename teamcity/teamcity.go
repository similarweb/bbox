package teamcity

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"
)

type Client struct {
	baseUrl   url.URL
	client    *http.Client
	BasicAuth *BasicAuth

	common service
	// Services of Teamcity
	Artifacts *ArtifactsService
	Queue     *QueueService
	Build     *BuildService
}

type BasicAuth struct {
	username string
	password string
}

type service struct {
	client *Client
}

// NewTeamCityClient creates a new TeamCity client.
func NewTeamCityClient(baseUrl url.URL, username, password string) *Client {
	newClient := &Client{
		baseUrl: baseUrl,
		BasicAuth: &BasicAuth{
			username: username,
			password: password,
		},
		client: &http.Client{},
	}

	newClient.initializeServices()

	return newClient
}

func (c *Client) initializeServices() {
	c.common.client = c
	c.Artifacts = (*ArtifactsService)(&c.common)
	c.Queue = (*QueueService)(&c.common)
	c.Build = (*BuildService)(&c.common)
}

// RequestOption represents an option that can modify an http.Request.
type RequestOption func(req *http.Request)

// NewRequestWrapper creates an API request with basic Auth with Username and Password.
// This Function injects the Accept and Content-Type headers.
// A relative URL can be provided in urlStr, in which case it is resolved relative to the BaseURL of the Client.
func (c *Client) NewRequestWrapper(method, urlStr string, body interface{}, opts ...RequestOption) (*http.Request, error) {
	if !strings.HasSuffix(c.baseUrl.Path, "/") {
		return nil, fmt.Errorf("BaseURL must have a trailing slash, but %v does not", c.baseUrl)
	}

	u, err := c.baseUrl.Parse(urlStr)
	if err != nil {
		return nil, err
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
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
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
