package teamcity

import (
	"fmt"
	"io"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type QueueService service

// ClearQueue cancels all queued builds in TeamCity using the REST API.
func (qs *QueueService) ClearQueue() error {
	req, err := qs.client.NewRequestWrapper("DELETE", "app/rest/buildQueue", nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	response, err := qs.client.client.Do(req)
	if err != nil {
		return fmt.Errorf("error executing request to clear the queue: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Errorf("error closing response body: %s", err)
		}
	}(response.Body)

	// Check the response. Expecting HTTP Status No Content (204) or OK (200) as a success indicator.
	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusNoContent {
		bodyBytes, _ := io.ReadAll(response.Body)
		return fmt.Errorf("failed to clear the queue, status code: %d, response: %s", response.StatusCode, string(bodyBytes))
	}

	return nil
}
