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
				WebURL:      "https://teamcity-example.com/",
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
				WebURL:      "https://teamcity-example.com/",
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
				WebURL:      "https://teamcity-example.com/",
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
			mockBuild := new(testutils.MockBuildService)
			mockArtifacts := new(testutils.MockArtifactsService)

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
