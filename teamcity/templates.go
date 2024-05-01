package teamcity

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/alitto/pond"
)

type VcsRootFromTemplateResponse struct {
	VCSRootEntries []struct {
		ID string `json:"id"`
	} `json:"vcs-root-entry"`
}

const (
	templatePondWorkerPoolSize   = 50
	templatePondChannelTasksSize = 100
)

type TemplateService service

// GetVcsRootsIDsFromTemplates retrieves VCS Root IDs from given template IDs.
func (vcs *VcsRootsService) GetVcsRootsIDsFromTemplates(templateIDs []string) ([]string, error) {
	vcsRootsIDs := []string{}

	for _, templateID := range templateIDs {
		vcsRootURL := fmt.Sprintf("app/rest/buildTypes/id:%s/vcs-root-entries?fields=vcs-root-entry", templateID)
		req, err := vcs.client.NewRequestWrapper("GET", vcsRootURL, nil)
		if err != nil {
			return []string{}, fmt.Errorf("error creating request: %v", err)
		}

		response, err := vcs.client.client.Do(req)
		if err != nil {
			return []string{}, fmt.Errorf("error executing request to get VCS Roots: %v", err)
		}
		defer response.Body.Close()

		if response.StatusCode != http.StatusOK {
			return []string{}, fmt.Errorf("failed to get VCS Roots, status code: %d", response.StatusCode)
		}

		var vcsRootResponse VcsRootFromTemplateResponse
		err = json.NewDecoder(response.Body).Decode(&vcsRootResponse)
		if err != nil {
			return []string{}, fmt.Errorf("error decoding response body: %v", err)
		}

		for _, vcsRootEntry := range vcsRootResponse.VCSRootEntries {
			vcsRootsIDs = append(vcsRootsIDs, vcsRootEntry.ID)
		}
	}
	return vcsRootsIDs, nil
}

// GetAllVcsRootsTemplates collects all VCS Roots IDs from a list of all project templates.
func (vcs *VcsRootsService) GetAllVcsRootsTemplates(allProjects []string) ([]string, error) {
	// Create a worker pool to concurrently fetch VCS Roots IDs from templates.
	pool := pond.New(templatePondWorkerPoolSize, templatePondWorkerPoolSize)
	defer pool.StopAndWait()

	var mu sync.Mutex // Protects unusedCount during concurrent increments.

	vcsRootsTemplates := []string{}
	var overallError error

	for _, projectID := range allProjects {
		localScopeProjectID := projectID
		pool.Submit(func() {
			if overallError != nil {
				return
			}
			templateIDs, err := vcs.GetProjectTemplates(localScopeProjectID)
			if err != nil {
				overallError = err
				return
			}

			templateVCSRootIDs, err := vcs.GetVcsRootsIDsFromTemplates(templateIDs)
			if err != nil {
				overallError = err
				return
			}

			mu.Lock()
			vcsRootsTemplates = append(vcsRootsTemplates, templateVCSRootIDs...)
			mu.Unlock()
		})
	}

	// Wait for all tasks to complete
	pool.StopAndWait()

	if overallError != nil {
		return nil, overallError
	}

	return vcsRootsTemplates, nil
}
