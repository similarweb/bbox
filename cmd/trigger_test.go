package cmd

import (
	"bbox/pkg/types"
	"bbox/teamcity"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
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
	// call GetBuildStatus to simulate the build finishing
	m.GetBuildStatus(buildNumber)
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
	m.GetArtifactChildren(buildID)
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
	m.GetAllBuildTypeArtifacts(buildID, buildTypeID)
	return args.Error(0)
}

func TestTrigger(t *testing.T) {
	tests := []struct {
		name                string
		buildTypeID         string
		branchName          string
		properties          map[string]string
		expectedTrigger     types.TriggerBuildWithParametersResponse
		waitForBuild        bool
		expectedWait        types.BuildStatusResponse
		expectArtifactErr   error
		expectedBuildErr    error
		buildHasArtifact    bool
		artifactsPath       string
		requireArtifacts    bool
		downloadArtifacts   bool
		waitForBuildTimeout time.Duration
	}{
		{
			name:        "Successful Trigger without Wait",
			buildTypeID: "bt123",
			branchName:  "master",
			properties:  map[string]string{"key": "value"},
			expectedTrigger: types.TriggerBuildWithParametersResponse{
				BuildTypeID: "bt123",
				WebURL:      "https://example.com/buildStatus",
				ID:          123,
				BuildType: types.BuildType{
					Name: "buildName",
				},
			},
			waitForBuild:        false,
			expectedBuildErr:    nil,
			buildHasArtifact:    false,
			artifactsPath:       "artifacts/",
			requireArtifacts:    false,
			downloadArtifacts:   false,
			waitForBuildTimeout: 15 * time.Minute,
		},
		{
			name:        "Successful Trigger with wait and no download",
			buildTypeID: "bt123",
			branchName:  "master",
			properties:  map[string]string{"key": "value"},
			expectedTrigger: types.TriggerBuildWithParametersResponse{
				BuildTypeID: "bt123",
				WebURL:      "https://example.com/buildStatus",
				ID:          123,
				BuildType: types.BuildType{
					Name: "buildName",
				},
			},
			waitForBuild:        true,
			expectedBuildErr:    nil,
			buildHasArtifact:    false,
			artifactsPath:       "artifacts/",
			requireArtifacts:    false,
			downloadArtifacts:   false,
			waitForBuildTimeout: 15 * time.Minute,
		},
		{
			name:        "Successful Trigger with wait and download",
			buildTypeID: "bt123",
			branchName:  "master",
			properties:  map[string]string{"key": "value"},
			expectedTrigger: types.TriggerBuildWithParametersResponse{
				BuildTypeID: "bt123",
				WebURL:      "https://example.com/buildStatus",
				ID:          123,
				BuildType: types.BuildType{
					Name: "buildName",
				},
			},
			expectedWait:        types.BuildStatusResponse{ID: 123, Status: "SUCCESS", State: "finished"},
			waitForBuild:        true,
			expectedBuildErr:    nil,
			buildHasArtifact:    true,
			artifactsPath:       "artifacts/",
			requireArtifacts:    false,
			downloadArtifacts:   true,
			waitForBuildTimeout: 15 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBuild := new(MockBuildService)
			mockArtifacts := new(MockArtifactsService)

			client := &teamcity.Client{
				Build:     mockBuild,
				Artifacts: mockArtifacts,
			}

			mockBuild.On("TriggerBuild", tt.buildTypeID, tt.branchName, tt.properties).Return(tt.expectedTrigger, tt.expectedBuildErr)

			if tt.waitForBuild {
				mockBuild.On("WaitForBuild", tt.expectedTrigger.BuildType.Name, tt.expectedTrigger.ID, 15*time.Minute).Return(tt.expectedWait, tt.expectedBuildErr)
				mockBuild.On("GetBuildStatus", tt.expectedTrigger.ID).Return(tt.expectedWait, tt.expectedBuildErr)
			}

			if tt.waitForBuild && tt.downloadArtifacts {
				mockArtifacts.On("BuildHasArtifact", tt.expectedTrigger.ID).Return(tt.buildHasArtifact)
				mockArtifacts.On("DownloadAndUnzipArtifacts", tt.expectedTrigger.ID, tt.buildTypeID, "artifacts/").Return(tt.expectArtifactErr)
				mockArtifacts.On("GetAllBuildTypeArtifacts", tt.expectedTrigger.ID, tt.buildTypeID).Return([]byte{}, tt.expectArtifactErr)
				mockArtifacts.On("GetArtifactChildren", tt.expectedTrigger.ID).Return(types.ArtifactChildren{}, tt.expectArtifactErr)
			}

			// Call the trigger function
			trigger(client, tt.buildTypeID, tt.branchName, tt.artifactsPath, tt.properties, tt.requireArtifacts, tt.waitForBuild, tt.downloadArtifacts, tt.waitForBuildTimeout)

			mockBuild.AssertExpectations(t)
			mockArtifacts.AssertExpectations(t)
		})
	}
}
