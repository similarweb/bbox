package teamcity

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/alitto/pond"

	log "github.com/sirupsen/logrus"
)

type VCSRootFromTemplateResponse struct {
	VCSRootEntries []struct {
		ID string `json:"id"`
	} `json:"vcs-root-entry"`
}

type TemplateService service

const workerPoolSize = 50

// GetVCSRootIDsFromTemplates retrieves VCS root IDs from given template IDs.
func (vcs *VCSRootService) GetVCSRootIDsFromTemplates(templateIDs []string) ([]string, error) {
	vcsRootIDs := []string{}

	for _, templateID := range templateIDs {
		vcsRootURL := fmt.Sprintf("app/rest/buildTypes/id:%s/vcs-root-entries?fields=vcs-root-entry", templateID)
		req, err := vcs.client.NewRequestWrapper("GET", vcsRootURL, nil)
		if err != nil {
			log.Errorf("error creating request: %v", err)
		}

		response, err := vcs.client.client.Do(req)
		if err != nil {
			log.Errorf("error executing request to get VCS roots: %v", err)
		}
		defer response.Body.Close()

		if response.StatusCode != http.StatusOK {
			log.Errorf("failed to get VCS roots, status code: %d", response.StatusCode)
		}

		var vcsRootResponse VCSRootFromTemplateResponse
		err = json.NewDecoder(response.Body).Decode(&vcsRootResponse)
		if err != nil {
			log.Errorf("error decoding response body: %v", err)
		}

		for _, vcsRootEntry := range vcsRootResponse.VCSRootEntries {
			vcsRootIDs = append(vcsRootIDs, vcsRootEntry.ID)
		}
	}
	return vcsRootIDs, nil
}

// GetAllVCSRootsTemplates collects all VCS root IDs from a list of all project templates.
func (vcs *VCSRootService) GetAllVCSRootsTemplates() ([]string, error) {
	projectIDs, err := vcs.GetAllProjects()
	if err != nil {
		return nil, err
	}

	pool := pond.New(workerPoolSize, 1000) // Create a pond with 50 workers and a buffered channel of 1000 tasks.
	defer pool.StopAndWait()

	var mu sync.Mutex // Protects unusedCount during concurrent increments.

	vcsRootTemplates := []string{}
	var overallError error

	for _, projectID := range projectIDs {
		projectID := projectID // avoid closure capturing issues
		pool.Submit(func() {
			if overallError != nil {
				return
			}
			templateIDs, err := vcs.GetProjectTemplates(projectID)
			if err != nil {
				overallError = err
				return
			}

			templateVCSRootIDs, err := vcs.GetVCSRootIDsFromTemplates(templateIDs)
			if err != nil {
				overallError = err
				return
			}

			mu.Lock()
			vcsRootTemplates = append(vcsRootTemplates, templateVCSRootIDs...)
			mu.Unlock()
		})
	}

	// Wait for all tasks to complete
	pool.StopAndWait()

	if overallError != nil {
		return nil, overallError
	}

	return vcsRootTemplates, nil
}
