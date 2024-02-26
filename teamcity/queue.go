package teamcity

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
)

// ClearQueue cancels all queued builds in TeamCity using the REST API.
func (tcc *Client) ClearQueue() error {

	reqURL := fmt.Sprintf("%s/app/rest/buildQueue", tcc.baseUrl)

	log.WithField("reqURL", reqURL).Debug("Clearing TeamCity build queue")

	req, err := http.NewRequest("DELETE", reqURL, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	req.SetBasicAuth(tcc.username, tcc.password)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Perform the HTTP DELETE request
	response, err := tcc.client.Do(req)
	if err != nil {
		return fmt.Errorf("error executing request to clear the queue: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Errorf("error closing response body: %s", err)
		}
	}(response.Body)

	// Check the response. Expecting HTTP Status No Content (204) or OK (200) as a success indicator.
	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusNoContent {
		bodyBytes, _ := ioutil.ReadAll(response.Body)
		return fmt.Errorf("failed to clear the queue, status code: %d, response: %s", response.StatusCode, string(bodyBytes))
	}

	return nil
}
