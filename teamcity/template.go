package teamcity

import (
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type VCSRootFromTemplateResponse struct {
	VCSRootEntries []struct {
		ID string `json:"id"`
	} `json:"vcs-root-entry"`
}

type TemplateService service

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

func (vcs *VCSRootService) GetAllVCSRootsTemplates() ([]string, error) {
	projectIDs, err := vcs.GetAllProjects()
	if err != nil {
		return nil, err
	}

	vcsRootTemplates := []string{}
	for _, projectID := range projectIDs {
		templateIDs, err := vcs.GetProjectTemplates(projectID)
		if err != nil {
			return nil, err
		}

		templateVCSRootIDs, err := vcs.GetVCSRootIDsFromTemplates(templateIDs)
		if err != nil {
			return nil, err
		}

		vcsRootTemplates = append(vcsRootTemplates, templateVCSRootIDs...)

	}

	return vcsRootTemplates, nil
}
