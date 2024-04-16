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

// getAllProjects retrieves all project IDs available in TeamCity.
func (vcs *VcsRootsService) GetAllProjects() ([]string, error) {
	req, err := vcs.client.NewRequestWrapper("GET", "app/rest/projects", nil)
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

	var allProjects ProjectsResponse
	err = json.NewDecoder(response.Body).Decode(&allProjects)
	if err != nil {
		log.Errorf("error decoding response body: %v", err)
	}

	projectIDs := []string{}
	for _, project := range allProjects.Projects {
		projectIDs = append(projectIDs, project.ID)
	}

	return projectIDs, nil
}

// GetProjectTemplates retrieves all template IDs associated with a given project ID.
func (vcs *VcsRootsService) GetProjectTemplates(projectID string) ([]string, error) {

	templatesURL := fmt.Sprintf("app/rest/projects/id:%s/templates", projectID)
	req, err := vcs.client.NewRequestWrapper("GET", templatesURL, nil)
	if err != nil {
		return []string{}, fmt.Errorf("error creating request: %w", err)
	}

	response, err := vcs.client.client.Do(req)
	if err != nil {
		return []string{}, fmt.Errorf("error executing request to get templates: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return []string{}, fmt.Errorf("failed to get templates, status code: %d", response.StatusCode)
	}

	var projectTemplates TemplateResponse
	err = json.NewDecoder(response.Body).Decode(&projectTemplates)
	if err != nil {
		return []string{}, fmt.Errorf("error decoding response body: %w", err)
	}

	templateIDs := []string{}
	for _, template := range projectTemplates.Templates {
		templateIDs = append(templateIDs, template.ID)
	}
	return templateIDs, nil
}
