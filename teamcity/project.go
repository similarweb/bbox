package teamcity

import (
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type ProjectsResponse struct {
	Projects []struct {
		ID string `json:"id"`
	} `json:"project"`
}

type TemplateResponse struct {
	Count     int `json:"count"`
	Templates []struct {
		ID string `json:"id"`
	} `json:"buildType"`
}

type ProjectService service

func (vcs *VCSRootService) GetAllProjects() ([]string, error) {
	projectsURL := "app/rest/projects"
	req, err := vcs.client.NewRequestWrapper("GET", projectsURL, nil)
	if err != nil {
		log.Errorf("error creating request: %v", err)
		return nil, err
	}

	response, err := vcs.client.client.Do(req)
	if err != nil {
		log.Errorf("error executing request to get projects: %v", err)
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		log.Errorf("failed to get projects, status code: %d", response.StatusCode)
	}

	var projectsResponse ProjectsResponse
	err = json.NewDecoder(response.Body).Decode(&projectsResponse)
	if err != nil {
		log.Errorf("error decoding response body: %v", err)
	}

	projectIDs := []string{}
	for _, project := range projectsResponse.Projects {
		projectIDs = append(projectIDs, project.ID)
	}

	fmt.Printf("Number of projects: %d\n", len(projectIDs))
	return projectIDs, nil
}

func (vcs *VCSRootService) GetProjectTemplates(projectID string) ([]string, error) {

	templatesURL := fmt.Sprintf("app/rest/projects/id:%s/templates", projectID)
	req, err := vcs.client.NewRequestWrapper("GET", templatesURL, nil)
	if err != nil {
		log.Errorf("error creating request: %v", err)
	}

	response, err := vcs.client.client.Do(req)
	if err != nil {
		log.Errorf("error executing request to get templates: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		log.Errorf("failed to get templates, status code: %d, id: %v", response.StatusCode, projectID)
	}

	var templatesResponse TemplateResponse
	err = json.NewDecoder(response.Body).Decode(&templatesResponse)
	if err != nil {
		log.Errorf("error decoding response body: %v, id: %v", err, projectID)
	}

	if templatesResponse.Count == 0 {
		return []string{}, nil
	}

	templateIDs := []string{}
	for _, template := range templatesResponse.Templates {
		templateIDs = append(templateIDs, template.ID)
	}
	return templateIDs, nil
}
