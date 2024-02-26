package teamcity

import (
	"net/http"
)

type Client struct {
	baseUrl  string
	username string
	password string
	client   *http.Client
}

func NewTeamCityClient(baseUrl, username, password string) *Client {
	return &Client{
		baseUrl:  baseUrl,
		username: username,
		password: password,
		client:   &http.Client{},
	}
}
