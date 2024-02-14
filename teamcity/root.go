package teamcity

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

type TeamCityClient struct {
	base_url string
	username string
	password string
	client   *http.Client
}

type BuildStatus struct {
	ID          string `xml:"id"`
	BuildTypeID string `xml:"buildTypeId"`
	Number      string `xml:"number"`
	Status      string `xml:"status"`
	State       string `xml:"state"`
}

func NewTeamCityClient(base_url, username, password string) *TeamCityClient {
	return &TeamCityClient{
		base_url: base_url,
		username: username,
		password: password,
		client:   &http.Client{},
	}
}

func (tcc *TeamCityClient) GetBuildStatus(buildId string) (*BuildStatus, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s/%s", tcc.base_url, "httpAuth/app/rest/builds", buildId), nil)
	if err != nil {
		return &BuildStatus{}, err
	}
	req.SetBasicAuth(tcc.username, tcc.password)

	resp, err := tcc.client.Do(req)
	if err != nil {
		return &BuildStatus{}, err
	}
	defer resp.Body.Close()

	bs := new(BuildStatus) //assume BuildStatus is a struct that represents the build status
	err = json.NewDecoder(resp.Body).Decode(bs)

	if err != nil {
		return &BuildStatus{}, err
	}

	return bs, nil
}

// TriggerBuild Assumes buildID corresponds to the build you want to trigger
func (tcc *TeamCityClient) TriggerBuild(buildId string) error {
	url := fmt.Sprintf("%s/%s", tcc.base_url, "httpAuth/app/rest/buildQueue")

	// Define a payload if required
	// It should contain any necessary parameters to trigger the relevant build
	var jsonStr = []byte(`{"buildType":{"id":"` + buildId + `"}}`)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))

	if err != nil {
		return err
	}

	req.SetBasicAuth(tcc.username, tcc.password)
	req.Header.Set("Content-Type", "application/json")

	log.Info("Triggering build: ", buildId)
	resp, err := tcc.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Confirm the request was processed successfully
	if resp.StatusCode != http.StatusOK {
		log.Errorf("Unexpected HTTP status: %d", resp.StatusCode)
		return nil
	}

	return nil
}

func (tcc *TeamCityClient) TriggerAndWaitForBuild(buildId string) (*BuildStatus, error) {
	err := tcc.TriggerBuild(buildId)
	if err != nil {
		return nil, err
	}

	var status *BuildStatus
	for {
		status, err = tcc.GetBuildStatus(buildId)
		if err != nil {
			return nil, err
		}

		if status.State == "finished" {
			break
		}

		time.Sleep(10 * time.Second)
	}

	return status, nil
}

func (tcc *TeamCityClient) GetBuilds() (string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", tcc.base_url, "httpAuth/app/rest/builds/"), nil)
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(tcc.username, tcc.password)

	resp, err := tcc.client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
