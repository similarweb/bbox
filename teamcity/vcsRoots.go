package teamcity

import (
	"encoding/json"
	"fmt"
	"net/http"

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

// GetAllVCSRootIDs gets all VCS root IDs.

func (vcs *VCSRootService) GetAllVCSRootIDs() ([]VCSRoot, error) {
	// vcs.GetAllVCSRootIDs()
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

	fmt.Printf("Number of vcsRoots: %d\n", len(allVCSRoots))

	return allVCSRoots, nil

}

func (vcs *VCSRootService) GetUnusedVCSRoots() error {
	// vcs.GetAllVCSRootIDs()
	allVCSRoots, err := vcs.GetAllVCSRootIDs()
	if err != nil {
		log.Errorf("error getting all VCS root IDs: %v", err)
	}

	vcsRootTemplats, err := vcs.GetAllVCSRootsTemplates()
	if err != nil {
		log.Errorf("error getting all VCS root templates: %v", err)
	}

	unusedCount := 0 // Counter for unused VCS roots
	for _, vcsRoot := range allVCSRoots {
		isUnused, err := vcs.IsVcsRootHaveInstance(vcsRoot.ID)
		if err != nil {
			log.Errorf("error checking if VCS root is unused: %v", err)
		}

		if isUnused {
			isInTemplate := false
			for _, templateID := range vcsRootTemplats {
				if vcsRoot.ID == templateID {
					isInTemplate = true
					break
				}
			}

			if !isInTemplate {
				unusedCount++ // if VCS root dosent have an instance and not in template
			}
		}
	}

	fmt.Printf("Number of unused VCS roots: %d\n", unusedCount)
	return nil
}

// Checks if the VCS root have instance.
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

	// Check VCS root instances
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
