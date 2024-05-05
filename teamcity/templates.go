package teamcity

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// todo - move types

type VcsRootFromTemplateResponse struct {
	VCSRootEntries []struct {
		ID string `json:"id"`
	} `json:"vcs-root-entry"`
}

type ITemplateService interface {
	GetVcsRootsIDsFromTemplates(templateIDs []string) ([]string, error)
}

type TemplateService service

// GetVcsRootsIDsFromTemplates retrieves VCS Root IDs from given template IDs.
func (template *TemplateService) GetVcsRootsIDsFromTemplates(templateIDs []string) ([]string, error) {
	vcsRootsIDs := []string{}

	for _, templateID := range templateIDs {
		vcsRootURL := fmt.Sprintf("app/rest/buildTypes/id:%s/vcs-root-entries?fields=vcs-root-entry", templateID)
		req, err := template.client.NewRequestWrapper("GET", vcsRootURL, nil)
		if err != nil {
			return []string{}, fmt.Errorf("error creating request: %w", err)
		}

		response, err := template.client.client.Do(req)
		if err != nil {
			return []string{}, fmt.Errorf("error executing request to get VCS Roots: %w", err)
		}
		defer response.Body.Close()

		if response.StatusCode != http.StatusOK {
			return []string{}, fmt.Errorf("failed to get VCS Roots, status code: %d", response.StatusCode)
		}

		var vcsRootResponse VcsRootFromTemplateResponse
		err = json.NewDecoder(response.Body).Decode(&vcsRootResponse)
		if err != nil {
			return []string{}, fmt.Errorf("error decoding response body: %w", err)
		}

		for _, vcsRootEntry := range vcsRootResponse.VCSRootEntries {
			vcsRootsIDs = append(vcsRootsIDs, vcsRootEntry.ID)
		}
	}
	return vcsRootsIDs, nil
}
