package multitrigger

import (
	"bbox/pkg/types"
	"bbox/pkg/utils/testutils"
	"bbox/teamcity"
	"testing"
	"time"
)

type buildTestCase struct {
	parameters        types.BuildParameters
	expectedTrigger   types.TriggerBuildWithParametersResponse
	triggerError      error
	triggerShouldFail bool
	waitError         error
	expectedWait      types.BuildStatusResponse
	waitShouldFail    bool
}

func TestTriggerBuilds(t *testing.T) {
	newTests := []struct {
		name               string
		buildsTriggered    []buildTestCase
		waitForBuilds      bool
		waitTimeout        time.Duration
		multiArtifactsPath string
		requireArtifacts   bool
		expectedResults    []types.BuildResult
	}{
		{
			name:               "Single Build with Artifacts",
			waitForBuilds:      true,
			waitTimeout:        30 * time.Second,
			multiArtifactsPath: "artifacts/",
			requireArtifacts:   true,
			expectedResults:    []types.BuildResult{},
			buildsTriggered: []buildTestCase{
				{
					parameters: types.BuildParameters{
						BuildTypeID:       "bt123",
						BranchName:        "master",
						PropertiesFlag:    map[string]string{"key": "value"},
						DownloadArtifacts: true,
					},
					expectedTrigger: types.TriggerBuildWithParametersResponse{
						BuildTypeID: "bt123",
						WebURL:      "https://example.com/buildStatus",
						ID:          123,
						BuildType: types.BuildType{
							Name: "buildName",
						},
					},
					triggerShouldFail: false,
					expectedWait:      types.BuildStatusResponse{ID: 123, Status: "SUCCESS", State: "finished"},
					waitShouldFail:    false,
				},
				{
					parameters: types.BuildParameters{
						BuildTypeID:       "bt1234",
						BranchName:        "master",
						PropertiesFlag:    map[string]string{"key": "value"},
						DownloadArtifacts: false,
					},
					expectedTrigger: types.TriggerBuildWithParametersResponse{
						BuildTypeID: "bt1234",
						WebURL:      "https://example.com/buildStatus",
						ID:          1234,
						BuildType: types.BuildType{
							Name: "buildName1",
						},
					},
					triggerShouldFail: false,
					expectedWait:      types.BuildStatusResponse{ID: 1234, Status: "SUCCESS", State: "finished"},
					waitShouldFail:    false,
				},
			},
		},
		{
			name:               "Two builds with artifacts required - one failed",
			waitForBuilds:      true,
			waitTimeout:        30 * time.Second,
			multiArtifactsPath: "artifacts/",
			requireArtifacts:   true,
			expectedResults:    []types.BuildResult{},
			buildsTriggered: []buildTestCase{
				{
					parameters: types.BuildParameters{
						BuildTypeID:       "bt123",
						BranchName:        "master",
						PropertiesFlag:    map[string]string{"key": "value"},
						DownloadArtifacts: true,
					},
					expectedTrigger: types.TriggerBuildWithParametersResponse{
						BuildTypeID: "bt123",
						WebURL:      "https://example.com/buildStatus",
						ID:          123,
						BuildType: types.BuildType{
							Name: "failedBuild",
						},
					},
					triggerShouldFail: false,
					expectedWait:      types.BuildStatusResponse{ID: 123, Status: "FAILURE", State: "finished"},
					waitShouldFail:    true,
				},
				{
					parameters: types.BuildParameters{
						BuildTypeID:       "bt1234",
						BranchName:        "master",
						PropertiesFlag:    map[string]string{"key": "value"},
						DownloadArtifacts: false,
					},
					expectedTrigger: types.TriggerBuildWithParametersResponse{
						BuildTypeID: "bt1234",
						WebURL:      "https://example.com/buildStatus",
						ID:          1234,
						BuildType: types.BuildType{
							Name: "buildName1",
						},
					},
					triggerShouldFail: false,
					expectedWait:      types.BuildStatusResponse{ID: 1234, Status: "SUCCESS", State: "finished"},
					waitShouldFail:    false,
				},
			},
		},
	}

	for _, tc := range newTests {
		t.Run(tc.name, func(t *testing.T) {
			mockBuildService := new(testutils.MockBuildService)
			mockArtifactsService := new(testutils.MockArtifactsService)
			client := &teamcity.Client{
				Build:     mockBuildService,
				Artifacts: mockArtifactsService,
			}

			var parameters []types.BuildParameters

			// Mocking the Build service
			for _, build := range tc.buildsTriggered {
				parameters = append(parameters, build.parameters)
				mockBuildService.On("TriggerBuild", build.parameters.BuildTypeID, build.parameters.BranchName, build.parameters.PropertiesFlag).Return(build.expectedTrigger, build.triggerError)
				if tc.waitForBuilds {
					mockBuildService.On("WaitForBuild", build.expectedTrigger.BuildType.Name, build.expectedTrigger.ID, tc.waitTimeout).Return(build.expectedWait, build.waitError)
					mockBuildService.On("GetBuildStatus", build.expectedTrigger.ID).Return(types.BuildStatusResponse{Status: "SUCCESS", State: "finished"}, nil)
				}
				if !build.waitShouldFail && build.parameters.DownloadArtifacts {
					mockArtifactsService.On("BuildHasArtifact", build.expectedTrigger.ID).Return(true)
					mockArtifactsService.On("DownloadAndUnzipArtifacts", build.expectedTrigger.ID, build.parameters.BuildTypeID, tc.multiArtifactsPath).Return(nil)
					mockArtifactsService.On("GetArtifactChildren", build.expectedTrigger.ID).Return(types.ArtifactChildren{}, nil)
					mockArtifactsService.On("GetAllBuildTypeArtifacts", build.expectedTrigger.ID, build.parameters.BuildTypeID).Return([]byte{}, nil)
				}
			}

			// Call the function
			triggerBuilds(client, parameters, tc.waitForBuilds, tc.waitTimeout, tc.multiArtifactsPath, tc.requireArtifacts)

			// Verify all expectations
			mockBuildService.AssertExpectations(t)
			mockArtifactsService.AssertExpectations(t)
		})

	}
}
