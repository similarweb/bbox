package teamcity

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/alitto/pond"
	log "github.com/sirupsen/logrus"
)

type VCSRoot struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Href string `json:"href"`
}

type VCSRootResponse struct {
	Count    int       `json:"count"`
	Href     string    `json:"href"`
	VCSRoot  []VCSRoot `json:"vcs-root"`
	NextHref string    `json:"nextHref"`
}

type VCSRootInstance struct {
	Count int `json:"count"`
}

type VCSRootService service

const pondWorkerPoolSize = 50
const pondChannelTasksSize = 1000

// GetAllVCSRootIDs retrieves all VCS root IDs, using pagination.
func (vcs *VCSRootService) GetAllVCSRootIDs() ([]VCSRoot, error) {
	allVCSRoots := []VCSRoot{}
	nextURL := "app/rest/vcs-roots"

	for nextURL != "" {
		req, err := vcs.client.NewRequestWrapper("GET", nextURL, nil)
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

		var vcsRootResponse VCSRootResponse
		err = json.NewDecoder(response.Body).Decode(&vcsRootResponse)
		if err != nil {
			log.Errorf("error decoding response body: %v", err)
		}

		allVCSRoots = append(allVCSRoots, vcsRootResponse.VCSRoot...)
		nextURL = vcsRootResponse.NextHref
	}

	return allVCSRoots, nil

}

// GetUnusedVCSRoots calculates the number of unused VCS roots.
// "Unused" VCS Root refers to a VCS Root that is neither linked to any build configurations nor included in any build templates.
func (vcs *VCSRootService) GetUnusedVCSRoots() (int, error) {
	allVCSRoots, err := vcs.GetAllVCSRootIDs()
	if err != nil {
		log.Errorf("error getting all VCS root IDs: %v", err)
	}

	vcsRootTemplats, err := vcs.GetAllVCSRootsTemplates()
	if err != nil {
		log.Errorf("error getting all VCS root templates: %v", err)
	}

	unusedCount := 0
	pool := pond.New(pondWorkerPoolSize, pondChannelTasksSize) // Create a pond with 50 workers and a buffered channel of 1000 tasks.
	defer pool.StopAndWait()

	var mu sync.Mutex // Protects unusedCount during concurrent increments.

	for _, vcsRoot := range allVCSRoots {
		vcsRoot := vcsRoot // Local scope redeclaration for closure
		pool.Submit(func() {
			isUnused, err := vcs.IsVcsRootHaveInstance(vcsRoot.ID)
			if err != nil {
				log.Errorf("error checking if VCS root is unused: %v", err)
				return
			}
			// Check if the VCS root not used in template
			if isUnused {
				isInTemplate := false
				for _, templateID := range vcsRootTemplats {
					if vcsRoot.ID == templateID {
						isInTemplate = true
						break
					}
				}

				if !isInTemplate {
					mu.Lock()
					unusedCount++
					if vcs.DeleteVCSRoot(vcsRoot.ID) {
						log.Infof("%s has been deleted", vcsRoot.ID)
					}
					mu.Unlock()
				}
			}
		})
	}

	pool.StopAndWait()

	return unusedCount, nil
}

// IsVcsRootHaveInstance checks if a VCS root has an instance.
func (vcs *VCSRootService) IsVcsRootHaveInstance(vcsRootID string) (bool, error) {
	vcsRootURL := fmt.Sprintf("app/rest/vcs-roots/id:%s", vcsRootID)
	req, err := vcs.client.NewRequestWrapper("GET", vcsRootURL, nil)
	if err != nil {
		log.Errorf("error creating request: %v", err)
	}

	response, err := vcs.client.client.Do(req)
	if err != nil {
		log.Errorf("error executing request to check if VCS root is unused: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		log.Errorf("failed to check if VCS root is unused, status code: %d", response.StatusCode)
	}

	var unusedResponse struct {
		Count int `json:"count"`
	}
	err = json.NewDecoder(response.Body).Decode(&unusedResponse)
	if err != nil {
		log.Errorf("error decoding response body: %v", err)
	}

	if unusedResponse.Count != 0 {
		return false, nil // VCS root is used
	}

	// Get VCS root instances
	instancesURL := fmt.Sprintf("app/rest/vcs-root-instances?locator=vcsRoot:(id:%s)", vcsRootID)
	req, err = vcs.client.NewRequestWrapper("GET", instancesURL, nil)
	if err != nil {
		log.Errorf("error creating request: %v", err)
	}

	response, err = vcs.client.client.Do(req)
	if err != nil {
		log.Errorf("error executing request to get VCS root instances: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		log.Errorf("failed to get VCS root instances, status code: %d", response.StatusCode)
	}

	var instancesResponse VCSRootInstance
	err = json.NewDecoder(response.Body).Decode(&instancesResponse)
	if err != nil {
		log.Errorf("error decoding response body: %v", err)
	}

	return instancesResponse.Count == 0, nil
}

// DeleteVCSRoot removes a VCS root by its ID.
func (vcs *VCSRootService) DeleteVCSRoot(vcsRootID string) bool {
	vcsRootURL := fmt.Sprintf("app/rest/vcs-roots/%s", vcsRootID)
	req, err := vcs.client.NewRequestWrapper("DELETE", vcsRootURL, nil)
	if err != nil {
		log.Errorf("error creating request for %v: %v", vcsRootID, err)
		return false
	}

	response, err := vcs.client.client.Do(req)
	if err != nil {
		log.Errorf("error executing request to delete %v: %v", vcsRootID, err)
		return false
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusNoContent {
		log.Errorf("failed to delete %v. status code: %d", vcsRootID, response.StatusCode)
		return false
	}

	return true
}
