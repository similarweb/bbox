package testutils

type EndpointConfig struct {
	Path       string
	Method     string
	Response   interface{}
	StatusCode int
}

type MockServerConfig struct {
	Endpoints []EndpointConfig
}
