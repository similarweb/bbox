package testutils

import (
	"bbox/pkg/types"
	"bbox/teamcity"
	"github.com/stretchr/testify/mock"
	"net/http"
	"time"
)

// MockTeamCityClient implements the interfaces used by the Build and Artifacts services.
type MockTeamCityClient struct {
	mock.Mock
	Build     teamcity.IBuildService
	Artifacts teamcity.IArtifactsService
}

func (m *MockTeamCityClient) NewRequestWrapper(method, urlStr string, body interface{}, opts ...teamcity.RequestOption) (*http.Request, error) {
	args := m.Called(method, urlStr, body, opts)
	return args.Get(0).(*http.Request), args.Error(1)
}

type MockBuildService struct {
	mock.Mock
}

func (m *MockBuildService) WaitForBuild(buildName string, buildNumber int, timeout time.Duration) (types.BuildStatusResponse, error) {
	args := m.Called(buildName, buildNumber, timeout)

	_, err := m.GetBuildStatus(buildNumber)
	if err != nil {
		return types.BuildStatusResponse{}, err
	}
	return args.Get(0).(types.BuildStatusResponse), args.Error(1)
}

func (m *MockBuildService) TriggerBuild(buildTypeID, branchName string, params map[string]string) (types.TriggerBuildWithParametersResponse, error) {
	args := m.Called(buildTypeID, branchName, params)
	return args.Get(0).(types.TriggerBuildWithParametersResponse), args.Error(1)
}

func (m *MockBuildService) GetBuildStatus(buildID int) (types.BuildStatusResponse, error) {
	args := m.Called(buildID)
	return args.Get(0).(types.BuildStatusResponse), args.Error(1)
}

type MockArtifactsService struct {
	mock.Mock
}

func (m *MockArtifactsService) GetAllBuildTypeArtifacts(buildID int, buildTypeID string) ([]byte, error) {
	args := m.Called(buildID, buildTypeID)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockArtifactsService) BuildHasArtifact(buildID int) bool {
	args := m.Called(buildID)
	_, err := m.GetArtifactChildren(buildID)
	if err != nil {
		return false
	}
	return args.Bool(0)
}

func (m *MockArtifactsService) GetArtifactChildren(buildID int) (types.ArtifactChildren, error) {
	args := m.Called(buildID)
	return args.Get(0).(types.ArtifactChildren), args.Error(1)
}

func (m *MockArtifactsService) GetArtifactContentByPath(path string) ([]byte, error) {
	args := m.Called(path)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockArtifactsService) DownloadAndUnzipArtifacts(buildID int, buildTypeID, destPath string) error {
	args := m.Called(buildID, buildTypeID, destPath)
	_, err := m.GetAllBuildTypeArtifacts(buildID, buildTypeID)
	if err != nil {
		return err
	}
	return args.Error(0)
}
