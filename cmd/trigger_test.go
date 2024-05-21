package cmd

import (
	"bbox/pkg/types"
	"bbox/pkg/utils/testutils"
	"bbox/teamcity"
	"testing"
	"time"
)

func TestTrigger(t *testing.T) {
	tests := []struct {
		name                             string
		buildTypeID                      string
		branchName                       string
		properties                       map[string]string
		triggerBuildResponse             types.TriggerBuildWithParametersResponse
		waitForBuild                     bool
		expectedWait                     types.BuildStatusResponse
		waitForBuildError                error
		buildHasArtifactsResponse        bool
		downloadAndUnzipArtifactsErr     error
		artifactsPath                    string
		requireArtifacts                 bool
		downloadArtifacts                bool
		waitForBuildTimeout              time.Duration
		getAllBuildTypeArtifactsResponse []byte
		getAllBuildTypeArtifactsError    error
		getArtifactChildrenResponse      types.ArtifactChildren
		getArtifactChildrenError         error
	}{
		{
			name:        "Successful Trigger without Wait",
			buildTypeID: "bt123",
			branchName:  "master",
			properties:  map[string]string{"key": "value"},
			triggerBuildResponse: types.TriggerBuildWithParametersResponse{
				BuildTypeID: "bt123",
				WebURL:      "https://teamcity-example.com/",
				ID:          123,
				BuildType: types.BuildType{
					Name: "buildName",
				},
			},
			waitForBuild:              false,
			waitForBuildError:         nil,
			buildHasArtifactsResponse: false,
			artifactsPath:             "artifacts/",
			requireArtifacts:          false,
			downloadArtifacts:         false,
			waitForBuildTimeout:       15 * time.Minute,
		},
		{
			name:        "Successful Trigger with wait and no download",
			buildTypeID: "bt123",
			branchName:  "master",
			properties:  map[string]string{"key": "value"},
			triggerBuildResponse: types.TriggerBuildWithParametersResponse{
				BuildTypeID: "bt123",
				WebURL:      "https://teamcity-example.com/",
				ID:          123,
				BuildType: types.BuildType{
					Name: "buildName",
				},
			},
			waitForBuild:              true,
			waitForBuildError:         nil,
			buildHasArtifactsResponse: false,
			artifactsPath:             "artifacts/",
			requireArtifacts:          false,
			downloadArtifacts:         false,
			waitForBuildTimeout:       15 * time.Minute,
		},
		{
			name:        "Successful Trigger with wait and download",
			buildTypeID: "bt123",
			branchName:  "master",
			properties:  map[string]string{"key": "value"},
			triggerBuildResponse: types.TriggerBuildWithParametersResponse{
				BuildTypeID: "bt123",
				WebURL:      "https://teamcity-example.com/",
				ID:          123,
				BuildType: types.BuildType{
					Name: "buildName",
				},
			},
			expectedWait:              types.BuildStatusResponse{ID: 123, Status: "SUCCESS", State: "finished"},
			waitForBuild:              true,
			waitForBuildError:         nil,
			buildHasArtifactsResponse: true,
			artifactsPath:             "artifacts/",
			requireArtifacts:          false,
			downloadArtifacts:         true,
			waitForBuildTimeout:       15 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBuild := new(testutils.MockBuildService)
			mockArtifacts := new(testutils.MockArtifactsService)

			client := &teamcity.Client{
				Build:     mockBuild,
				Artifacts: mockArtifacts,
			}

			mockBuild.On("TriggerBuild", tt.buildTypeID, tt.branchName, tt.properties).Return(tt.triggerBuildResponse, tt.waitForBuildError)

			if tt.waitForBuild {
				mockBuild.On("WaitForBuild", tt.triggerBuildResponse.BuildType.Name, tt.triggerBuildResponse.ID, tt.waitForBuildTimeout).Return(tt.expectedWait, tt.waitForBuildError)
				mockBuild.On("GetBuildStatus", tt.triggerBuildResponse.ID).Return(tt.expectedWait, tt.waitForBuildError)
			}

			if tt.waitForBuild && tt.downloadArtifacts {
				mockArtifacts.On("BuildHasArtifact", tt.triggerBuildResponse.ID).Return(tt.buildHasArtifactsResponse)
				mockArtifacts.On("DownloadAndUnzipArtifacts", tt.triggerBuildResponse.ID, tt.buildTypeID, tt.artifactsPath).Return(tt.downloadAndUnzipArtifactsErr)
				mockArtifacts.On("GetAllBuildTypeArtifacts", tt.triggerBuildResponse.ID, tt.buildTypeID).Return(tt.getAllBuildTypeArtifactsResponse, tt.getAllBuildTypeArtifactsError)
				mockArtifacts.On("GetArtifactChildren", tt.triggerBuildResponse.ID).Return(tt.getArtifactChildrenResponse, tt.getArtifactChildrenError)
			}

			trigger(client, tt.buildTypeID, tt.branchName, tt.artifactsPath, tt.properties, tt.requireArtifacts, tt.waitForBuild, tt.downloadArtifacts, tt.waitForBuildTimeout)

			mockBuild.AssertExpectations(t)
			mockArtifacts.AssertExpectations(t)
		})
	}
}
