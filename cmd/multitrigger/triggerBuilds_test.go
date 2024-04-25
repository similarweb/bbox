package multitrigger

import (
	"bbox/pkg/types"
	"bbox/pkg/utils/testutils"
	"bbox/teamcity"
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type buildTestCase struct {
	parameters                types.BuildParameters
	expectedTrigger           types.TriggerBuildWithParametersResponse
	triggerError              error
	triggerShouldFail         bool
	waitError                 error
	expectedWait              types.BuildStatusResponse
	waitShouldFail            bool
	downloadError             error
	expectedBuildHasArtifacts bool
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
		exitError          error
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
					triggerShouldFail:         false,
					expectedWait:              types.BuildStatusResponse{ID: 123, Status: "SUCCESS", State: "finished"},
					waitShouldFail:            false,
					expectedBuildHasArtifacts: true,
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
					triggerShouldFail:         false,
					expectedWait:              types.BuildStatusResponse{ID: 1234, Status: "SUCCESS", State: "finished"},
					waitShouldFail:            false,
					expectedBuildHasArtifacts: true,
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
		{
			name:               "build with trigger failure",
			waitForBuilds:      true,
			waitTimeout:        30 * time.Second,
			multiArtifactsPath: "artifacts/",
			requireArtifacts:   true,
			expectedResults:    []types.BuildResult{},
			exitError:          errors.New("error triggering build: this is a test error"),
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
							Name: "failedTrigger",
						},
					},
					triggerShouldFail: true,
					triggerError:      errors.New("this is a test error"),
					expectedWait:      types.BuildStatusResponse{ID: 123, Status: "FAILURE", State: "finished"},
					waitShouldFail:    true,
				},
			},
		},
		{
			name:               "build with wait failure",
			waitForBuilds:      true,
			waitTimeout:        30 * time.Second,
			multiArtifactsPath: "artifacts/",
			requireArtifacts:   true,
			expectedResults:    []types.BuildResult{},
			exitError:          errors.New("error waiting for build: this is a test error"),
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
							Name: "failedWait",
						},
					},
					triggerShouldFail: false,
					triggerError:      nil,
					waitError:         errors.New("this is a test error"),
					expectedWait:      types.BuildStatusResponse{},
					waitShouldFail:    true,
				},
			},
		},
		{
			name:               "download artifacts failure",
			waitForBuilds:      true,
			waitTimeout:        30 * time.Second,
			multiArtifactsPath: "artifacts/",
			requireArtifacts:   true,
			expectedResults:    []types.BuildResult{},
			exitError:          errors.New("error handling artifacts: error downloading artifacts: this is a test error"),
			buildsTriggered: []buildTestCase{
				{
					parameters: types.BuildParameters{
						BuildTypeID:       "bt123456",
						BranchName:        "master",
						PropertiesFlag:    map[string]string{"key": "value"},
						DownloadArtifacts: true,
					},
					expectedTrigger: types.TriggerBuildWithParametersResponse{
						BuildTypeID: "bt123",
						WebURL:      "https://example.com/buildStatus",
						ID:          123456,
						BuildType: types.BuildType{
							Name: "failedWait",
						},
					},
					triggerShouldFail:         false,
					triggerError:              nil,
					waitError:                 nil,
					expectedWait:              types.BuildStatusResponse{ID: 123456, Status: "SUCCESS", State: "finished"},
					waitShouldFail:            false,
					downloadError:             errors.New("this is a test error"),
					expectedBuildHasArtifacts: true,
				},
			},
		},
		{
			name:               "artifacts required but not found",
			waitForBuilds:      true,
			waitTimeout:        30 * time.Second,
			multiArtifactsPath: "artifacts/",
			requireArtifacts:   true,
			expectedResults:    []types.BuildResult{},
			exitError:          errors.New("error handling artifacts: build requires artifacts and did not produce any"),
			buildsTriggered: []buildTestCase{
				{
					parameters: types.BuildParameters{
						BuildTypeID:       "bt123456",
						BranchName:        "master",
						PropertiesFlag:    map[string]string{"key": "value"},
						DownloadArtifacts: true,
					},
					expectedTrigger: types.TriggerBuildWithParametersResponse{
						BuildTypeID: "bt123",
						WebURL:      "https://example.com/buildStatus",
						ID:          123456,
						BuildType: types.BuildType{
							Name: "failedWait",
						},
					},
					triggerShouldFail:         false,
					triggerError:              nil,
					waitError:                 nil,
					expectedWait:              types.BuildStatusResponse{ID: 123456, Status: "SUCCESS", State: "finished"},
					waitShouldFail:            false,
					downloadError:             errors.New("this is a test error"),
					expectedBuildHasArtifacts: false,
				},
			},
		},
		{
			name:               "artifacts not required and not found",
			waitForBuilds:      true,
			waitTimeout:        30 * time.Second,
			multiArtifactsPath: "artifacts/",
			requireArtifacts:   false,
			expectedResults:    []types.BuildResult{},
			//exitError:          errors.New("error handling artifacts: build requires artifacts and did not produce any"),
			buildsTriggered: []buildTestCase{
				{
					parameters: types.BuildParameters{
						BuildTypeID:       "bt123456",
						BranchName:        "master",
						PropertiesFlag:    map[string]string{"key": "value"},
						DownloadArtifacts: true,
					},
					expectedTrigger: types.TriggerBuildWithParametersResponse{
						BuildTypeID: "bt123",
						WebURL:      "https://example.com/buildStatus",
						ID:          123456,
						BuildType: types.BuildType{
							Name: "failedWait",
						},
					},
					triggerShouldFail: false,
					triggerError:      nil,
					waitError:         nil,
					expectedWait:      types.BuildStatusResponse{ID: 123456, Status: "SUCCESS", State: "finished"},
					waitShouldFail:    false,
					//downloadError:             errors.New("this is a test error"),
					expectedBuildHasArtifacts: false,
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
				if !build.triggerShouldFail && tc.waitForBuilds {
					mockBuildService.On("WaitForBuild", build.expectedTrigger.BuildType.Name, build.expectedTrigger.ID, tc.waitTimeout).Return(build.expectedWait, build.waitError)
					mockBuildService.On("GetBuildStatus", build.expectedTrigger.ID).Return(types.BuildStatusResponse{Status: "SUCCESS", State: "finished"}, nil)
				}
				if !build.waitShouldFail && build.parameters.DownloadArtifacts {
					mockArtifactsService.On("BuildHasArtifact", build.expectedTrigger.ID).Return(build.expectedBuildHasArtifacts)
					mockArtifactsService.On("GetArtifactChildren", build.expectedTrigger.ID).Return(types.ArtifactChildren{}, nil)
					if build.expectedBuildHasArtifacts {
						mockArtifactsService.On("DownloadAndUnzipArtifacts", build.expectedTrigger.ID, build.parameters.BuildTypeID, tc.multiArtifactsPath).Return(build.downloadError)
						mockArtifactsService.On("GetAllBuildTypeArtifacts", build.expectedTrigger.ID, build.parameters.BuildTypeID).Return([]byte{}, nil)
					}
				}
			}

			// Call the function
			err := triggerBuilds(client, parameters, tc.waitForBuilds, tc.waitTimeout, tc.multiArtifactsPath, tc.requireArtifacts)

			// assert error
			if tc.exitError != nil {
				assert.EqualError(t, err, tc.exitError.Error())
			} else {
				assert.NoError(t, err)

			}

			// Verify all expectations
			mockBuildService.AssertExpectations(t)
			mockArtifactsService.AssertExpectations(t)
		})

	}
}
