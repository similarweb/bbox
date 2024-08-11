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
	parameters                       types.BuildParameters
	triggerBuildResponse             types.TriggerBuildWithParametersResponse
	triggerBuildError                error
	triggerShouldFail                bool
	waitForBuildError                error
	waitForBuildResponse             types.BuildStatusResponse
	waitShouldFail                   bool
	downloadError                    error
	buildHasArtifactsResponse        bool
	getBuildStatusResponse           types.BuildStatusResponse
	getBuildStatusError              error
	getArtifactChildrenResponse      types.ArtifactChildren
	getArtifactChildrenError         error
	getAllBuildTypeArtifactsResponse []byte
	getAllBuildTypeArtifactsError    error
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
			name:               "Two Builds with Artifacts Required - Both Successful",
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
					triggerBuildResponse: types.TriggerBuildWithParametersResponse{
						BuildTypeID: "bt123",
						WebURL:      "https://teamcity-example.com/",
						ID:          123,
						BuildType: types.BuildType{
							Name: "buildName",
						},
					},
					triggerShouldFail:         false,
					waitForBuildResponse:      types.BuildStatusResponse{ID: 123, Status: "SUCCESS", State: "finished"},
					waitShouldFail:            false,
					buildHasArtifactsResponse: true,
				},
				{
					parameters: types.BuildParameters{
						BuildTypeID:       "bt1234",
						BranchName:        "master",
						PropertiesFlag:    map[string]string{"key": "value"},
						DownloadArtifacts: true,
					},
					triggerBuildResponse: types.TriggerBuildWithParametersResponse{
						BuildTypeID: "bt1234",
						WebURL:      "https://teamcity-example.com/",
						ID:          1234,
						BuildType: types.BuildType{
							Name: "buildName1",
						},
					},
					triggerShouldFail:         false,
					waitForBuildResponse:      types.BuildStatusResponse{ID: 1234, Status: "SUCCESS", State: "finished"},
					waitShouldFail:            false,
					buildHasArtifactsResponse: true,
				},
			},
		},
		{
			name:               "Two Builds With Artifacts Required - One Failed",
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
					triggerBuildResponse: types.TriggerBuildWithParametersResponse{
						BuildTypeID: "bt123",
						WebURL:      "https://teamcity-example.com/",
						ID:          123,
						BuildType: types.BuildType{
							Name: "failedBuild",
						},
					},
					waitForBuildError:    errors.New("this is a test error"),
					triggerShouldFail:    false,
					waitForBuildResponse: types.BuildStatusResponse{ID: 123, Status: "FAILURE", State: "finished"},
					waitShouldFail:       true,
				},
				{
					parameters: types.BuildParameters{
						BuildTypeID:       "bt1234",
						BranchName:        "master",
						PropertiesFlag:    map[string]string{"key": "value"},
						DownloadArtifacts: false,
					},
					triggerBuildResponse: types.TriggerBuildWithParametersResponse{
						BuildTypeID: "bt1234",
						WebURL:      "https://teamcity-example.com/",
						ID:          1234,
						BuildType: types.BuildType{
							Name: "buildName1",
						},
					},
					triggerShouldFail:    false,
					waitForBuildResponse: types.BuildStatusResponse{ID: 1234, Status: "SUCCESS", State: "finished"},
					waitShouldFail:       false,
				},
			},
		},
		{
			name:               "Build With Trigger Failure",
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
					triggerBuildResponse: types.TriggerBuildWithParametersResponse{
						BuildTypeID: "bt123",
						WebURL:      "https://teamcity-example.com/",
						ID:          123,
						BuildType: types.BuildType{
							Name: "failedTrigger",
						},
					},
					triggerShouldFail:    true,
					triggerBuildError:    errors.New("this is a test error"),
					waitForBuildResponse: types.BuildStatusResponse{ID: 123, Status: "FAILURE", State: "finished"},
					waitShouldFail:       true,
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
					triggerBuildResponse: types.TriggerBuildWithParametersResponse{
						BuildTypeID: "bt123",
						WebURL:      "https://teamcity-example.com/",
						ID:          123,
						BuildType: types.BuildType{
							Name: "failedWait",
						},
					},
					triggerShouldFail:    false,
					triggerBuildError:    nil,
					waitForBuildError:    errors.New("this is a test error"),
					waitForBuildResponse: types.BuildStatusResponse{},
					waitShouldFail:       true,
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
					triggerBuildResponse: types.TriggerBuildWithParametersResponse{
						BuildTypeID: "bt123",
						WebURL:      "https://teamcity-example.com/",
						ID:          123456,
						BuildType: types.BuildType{
							Name: "failedWait",
						},
					},
					triggerShouldFail:         false,
					triggerBuildError:         nil,
					waitForBuildError:         nil,
					waitForBuildResponse:      types.BuildStatusResponse{ID: 123456, Status: "SUCCESS", State: "finished"},
					waitShouldFail:            false,
					downloadError:             errors.New("this is a test error"),
					buildHasArtifactsResponse: true,
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
					triggerBuildResponse: types.TriggerBuildWithParametersResponse{
						BuildTypeID: "bt123",
						WebURL:      "https://teamcity-example.com/",
						ID:          123456,
						BuildType: types.BuildType{
							Name: "failedWait",
						},
					},
					triggerShouldFail:         false,
					triggerBuildError:         nil,
					waitForBuildError:         nil,
					waitForBuildResponse:      types.BuildStatusResponse{ID: 123456, Status: "SUCCESS", State: "finished"},
					waitShouldFail:            false,
					downloadError:             errors.New("this is a test error"),
					buildHasArtifactsResponse: false,
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
			buildsTriggered: []buildTestCase{
				{
					parameters: types.BuildParameters{
						BuildTypeID:       "bt123456",
						BranchName:        "master",
						PropertiesFlag:    map[string]string{"key": "value"},
						DownloadArtifacts: true,
					},
					triggerBuildResponse: types.TriggerBuildWithParametersResponse{
						BuildTypeID: "bt123",
						WebURL:      "https://teamcity-example.com/",
						ID:          123456,
						BuildType: types.BuildType{
							Name: "failedWait",
						},
					},
					triggerShouldFail:         false,
					triggerBuildError:         nil,
					waitForBuildError:         nil,
					waitForBuildResponse:      types.BuildStatusResponse{ID: 123456, Status: "SUCCESS", State: "finished"},
					waitShouldFail:            false,
					buildHasArtifactsResponse: false,
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

			for _, build := range tc.buildsTriggered {
				parameters = append(parameters, build.parameters)
				mockBuildService.On("TriggerBuild", build.parameters.BuildTypeID, build.parameters.BranchName, build.parameters.PropertiesFlag).Return(build.triggerBuildResponse, build.triggerBuildError)
				if !build.triggerShouldFail && tc.waitForBuilds {
					mockBuildService.On("WaitForBuild", build.triggerBuildResponse.BuildType.Name, build.triggerBuildResponse.ID, tc.waitTimeout).Return(build.waitForBuildResponse, build.waitForBuildError)
					mockBuildService.On("GetBuildStatus", build.triggerBuildResponse.ID).Return(build.getBuildStatusResponse, build.getBuildStatusError)
				}
				if !build.waitShouldFail && build.parameters.DownloadArtifacts {
					mockArtifactsService.On("BuildHasArtifact", build.triggerBuildResponse.ID).Return(build.buildHasArtifactsResponse)
					mockArtifactsService.On("GetArtifactChildren", build.triggerBuildResponse.ID).Return(build.getArtifactChildrenResponse, build.getArtifactChildrenError)
					if build.buildHasArtifactsResponse {
						mockArtifactsService.On("DownloadAndUnzipArtifacts", build.triggerBuildResponse.ID, build.parameters.BuildTypeID, tc.multiArtifactsPath).Return(build.downloadError)
						mockArtifactsService.On("GetAllBuildTypeArtifacts", build.triggerBuildResponse.ID, build.parameters.BuildTypeID).Return(build.getAllBuildTypeArtifactsResponse, build.getAllBuildTypeArtifactsError)
					}
				}
			}

			err := triggerBuilds(client, parameters, tc.waitForBuilds, tc.waitTimeout, tc.multiArtifactsPath, tc.requireArtifacts)

			if tc.exitError != nil {
				assert.EqualError(t, err, tc.exitError.Error())
			} else {
				assert.NoError(t, err)

			}

			mockBuildService.AssertExpectations(t)
			mockArtifactsService.AssertExpectations(t)
		})

	}
}
