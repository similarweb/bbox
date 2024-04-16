package teamcity

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/alitto/pond"
	log "github.com/sirupsen/logrus"
)

type VcsRoots struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Href string `json:"href"`
}

type VcsRootsResponse struct {
	Count    int        `json:"count"`
	Href     string     `json:"href"`
	VcsRoots []VcsRoots `json:"vcs-root"`
	NextHref string     `json:"nextHref"`
}

type VcsRootInstanceResponse struct {
	Count int `json:"count"`
}

type VcsRootsService service

const (
	vcsRootPondWorkerPoolSize   = 50
	vcsRootPondChannelTasksSize = 100
)

// GetAllVcsRootsIDs retrieves all vcs root IDs, using pagination.
func (vcs *VcsRootsService) GetAllVcsRootsIDs() ([]VcsRoots, error) {
	allVcsRoots := []VcsRoots{}
	nextURL := "app/rest/vcs-roots"

	for nextURL != "" {
		req, err := vcs.client.NewRequestWrapper("GET", nextURL, nil)
		if err != nil {
			return allVcsRoots, fmt.Errorf("error creating request: %v", err)
		}

		response, err := vcs.client.client.Do(req)
		if err != nil {
			return allVcsRoots, fmt.Errorf("error executing request to get vcs roots: %v", err)
		}
		defer response.Body.Close()

		if response.StatusCode != http.StatusOK {
			return allVcsRoots, fmt.Errorf("failed to get vcs roots, status code: %d", response.StatusCode)
		}

		var currentVcsRootsResponse VcsRootsResponse
		err = json.NewDecoder(response.Body).Decode(&currentVcsRootsResponse)
		if err != nil {
			return allVcsRoots, fmt.Errorf("error decoding response body: %v", err)
		}

		allVcsRoots = append(allVcsRoots, currentVcsRootsResponse.VcsRoots...)
		nextURL = currentVcsRootsResponse.NextHref
	}

	return allVcsRoots, nil
}

// GetUnusedVcsRootsIDs retrieves all unused vcs root IDs.
// Unused vcs root refers to a vcs root that is neither linked to any build configurations nor included in any build templates.
func (vcs *VcsRootsService) GetUnusedVcsRootsIDs(allVcsRoots []VcsRoots, allVcsRootsTemplates []string) ([]string, error) {
	unusedVcsRoots := []string{}
	pool := pond.New(vcsRootPondWorkerPoolSize, vcsRootPondChannelTasksSize)
	defer pool.StopAndWait()

	var mu sync.Mutex // Protects unusedVcsRoots slice during concurrent access.

	for _, vcsRoot := range allVcsRoots {
		localVcsRoot := vcsRoot // Local scope redeclaration for closure
		pool.Submit(func() {
			isUsed, err := vcs.IsVcsRootHaveInstance(localVcsRoot.ID)
			if err != nil {
				log.Errorf("error checking if %s has instances: %v", localVcsRoot.ID, err)
				return
			}

			if isUsed {
				isInTemplate := false
				for _, templateID := range allVcsRootsTemplates {
					if localVcsRoot.ID == templateID {
						isInTemplate = true
						break
					}
				}

				if !isInTemplate {
					mu.Lock()
					unusedVcsRoots = append(unusedVcsRoots, localVcsRoot.ID)
					mu.Unlock()
				}
			}
		})
	}

	pool.StopAndWait()
	return unusedVcsRoots, nil
}

// DeleteUnusedVcsRoots deletes all unused vcs roots.
func (vcs *VcsRootsService) DeleteUnusedVcsRoots(allUnusedVcsRoots []string) (int, error) {
	pool := pond.New(vcsRootPondWorkerPoolSize, vcsRootPondChannelTasksSize)
	defer pool.StopAndWait()

	for _, id := range allUnusedVcsRoots {
		localID := id // Local scope redeclaration for closure
		pool.Submit(func() {
			if deleted, err := vcs.DeleteVcsRoot(localID); err == nil && deleted {
				log.Infof("%s has been deleted", localID)
			} else {
				log.Errorf("error deleting %s: %v", localID, err)
			}
		})
	}

	pool.StopAndWait()
	return len(allUnusedVcsRoots), nil
}

// IsVcsRootHaveInstance checks if a vcs root has an instance.
func (vcs *VcsRootsService) IsVcsRootHaveInstance(vcsRootID string) (bool, error) {

	var instancesResponse VcsRootInstanceResponse
	// Get vcs root instances
	instancesURL := fmt.Sprintf("app/rest/vcs-root-instances?locator=vcsRoot:(id:%s)", vcsRootID)
	req, err := vcs.client.NewRequestWrapper("GET", instancesURL, nil)
	if err != nil {
		return false, fmt.Errorf("error creating request: %w", err)
	}

	response, err := vcs.client.client.Do(req)
	if err != nil {
		return false, fmt.Errorf("error executing request to get vcs root instances: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return false, fmt.Errorf("failed to get vcs root instances, status code: %d", response.StatusCode)
	}

	err = json.NewDecoder(response.Body).Decode(&instancesResponse)
	if err != nil {
		return false, fmt.Errorf("error decoding response body: %w", err)
	}

	return instancesResponse.Count == 0, nil // if count is 0, then vcs root has 0 uses as a instances
}

// DeleteVcsRoot removes a vcs root by its ID.
func (vcs *VcsRootsService) DeleteVcsRoot(vcsRootID string) (bool, error) {
	vcsRootURL := fmt.Sprintf("app/rest/vcs-roots/%s", vcsRootID)
	req, err := vcs.client.NewRequestWrapper("DELETE", vcsRootURL, nil)
	if err != nil {
		return false, fmt.Errorf("error creating request for %v: %v", vcsRootID, err)
	}

	response, err := vcs.client.client.Do(req)
	if err != nil {
		return false, fmt.Errorf("error executing request to delete %v: %v", vcsRootID, err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusNoContent {
		return false, fmt.Errorf("failed to delete %v. status code: %d", vcsRootID, response.StatusCode)
	}

	return true, nil
}
