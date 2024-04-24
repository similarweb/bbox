package multitrigger

import (
	"bbox/pkg/types"
	"bbox/pkg/utils/testutils"
	"bbox/teamcity"
	"testing"
	"time"
)

type buildTestCase struct {
	parameters types.BuildParameters
	//branch    string
	//BuildTypeID string
	//downloadArtifacts bool
	expectedTrigger types.TriggerBuildWithParametersResponse
	triggerError    error
	expectedWait    types.BuildStatusResponse
	waitError       error
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
					triggerError: nil,
					expectedWait: types.BuildStatusResponse{ID: 123, Status: "SUCCESS", State: "finished"},
					waitError:    nil,
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
					triggerError: nil,
					expectedWait: types.BuildStatusResponse{ID: 1234, Status: "SUCCESS", State: "finished"},
					waitError:    nil,
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
					triggerError: nil,
					// todo - make this a failed build
					expectedWait: types.BuildStatusResponse{ID: 123, Status: "SUCCESS", State: "finished"},
					waitError:    nil,
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
					triggerError: nil,
					expectedWait: types.BuildStatusResponse{ID: 1234, Status: "SUCCESS", State: "finished"},
					waitError:    nil,
				},
			},
		},
	}
	// Define the test cases
	//tests := []struct {
	//	name               string
	//	parameters         []types.BuildParameters
	//	waitForBuilds      bool
	//	waitTimeout        time.Duration
	//	multiArtifactsPath string
	//	requireArtifacts   bool
	//	expectedTrigger    types.TriggerBuildWithParametersResponse
	//	expectedWait       types.BuildStatusResponse
	//	waitError          error
	//	expectedResults    []types.BuildResult // Define what you expect to receive on the results channel
	//}{
	//	{
	//		name: "Single Build with Artifacts",
	//		parameters: []types.BuildParameters{
	//			{
	//				BuildTypeID:       "bt123",
	//				BranchName:        "master",
	//				PropertiesFlag:    map[string]string{"key": "value"},
	//				DownloadArtifacts: true,
	//			},
	//			{
	//				BuildTypeID:       "bt1234",
	//				BranchName:        "master",
	//				PropertiesFlag:    map[string]string{"key": "value"},
	//				DownloadArtifacts: true,
	//			},
	//			{
	//				BuildTypeID:       "bt12345",
	//				BranchName:        "master",
	//				PropertiesFlag:    map[string]string{"key": "value"},
	//				DownloadArtifacts: true,
	//			},
	//		},
	//		waitForBuilds:      true,
	//		waitTimeout:        30 * time.Second,
	//		multiArtifactsPath: "artifacts/",
	//		requireArtifacts:   true,
	//		expectedTrigger: types.TriggerBuildWithParametersResponse{
	//			BuildTypeID: "bt123",
	//			WebURL:      "https://example.com/buildStatus",
	//			ID:          123,
	//			BuildType: types.BuildType{
	//				Name: "buildName",
	//			},
	//		},
	//		expectedWait: types.BuildStatusResponse{ID: 123, Status: "SUCCESS", State: "finished"},
	//		waitError:    nil,
	//		expectedResults: []types.BuildResult{
	//			{
	//				BuildName:           "bt123",
	//				WebURL:              "https://example.com/buildStatus",
	//				BranchName:          "master",
	//				BuildStatus:         "SUCCESS",
	//				DownloadedArtifacts: true,
	//				Error:               nil,
	//			},
	//		},
	//	},
	//}

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
			for _, param := range tc.buildsTriggered {
				parameters = append(parameters, param.parameters)
				mockBuildService.On("TriggerBuild", param.parameters.BuildTypeID, param.parameters.BranchName, param.parameters.PropertiesFlag).Return(param.expectedTrigger, param.triggerError)
				if tc.waitForBuilds {
					mockBuildService.On("WaitForBuild", param.expectedTrigger.BuildType.Name, param.expectedTrigger.ID, tc.waitTimeout).Return(param.expectedWait, param.waitError)
					mockBuildService.On("GetBuildStatus", param.expectedTrigger.ID).Return(types.BuildStatusResponse{Status: "SUCCESS", State: "finished"}, nil)
				}
				if param.parameters.DownloadArtifacts {
					mockArtifactsService.On("BuildHasArtifact", param.expectedTrigger.ID).Return(true)
					mockArtifactsService.On("DownloadAndUnzipArtifacts", param.expectedTrigger.ID, param.parameters.BuildTypeID, tc.multiArtifactsPath).Return(nil)
					mockArtifactsService.On("GetArtifactChildren", param.expectedTrigger.ID).Return(types.ArtifactChildren{}, nil)
					mockArtifactsService.On("GetAllBuildTypeArtifacts", param.expectedTrigger.ID, param.parameters.BuildTypeID).Return([]byte{}, nil)
				}
			}

			// Call the function
			triggerBuilds(client, parameters, tc.waitForBuilds, tc.waitTimeout, tc.multiArtifactsPath, tc.requireArtifacts)

			// Verify all expectations
			mockBuildService.AssertExpectations(t)
			mockArtifactsService.AssertExpectations(t)
		})

	}

	//for _, tc := range tests {
	//	t.Run(tc.name, func(t *testing.T) {
	//		mockBuildService := new(testutils.MockBuildService)
	//		mockArtifactsService := new(testutils.MockArtifactsService)
	//		client := &teamcity.Client{
	//			Build:     mockBuildService,
	//			Artifacts: mockArtifactsService,
	//		}
	//
	//		// Mocking the Build service
	//		for _, param := range tc.parameters {
	//			mockBuildService.On("TriggerBuild", param.BuildTypeID, param.BranchName, param.PropertiesFlag).Return(tc.expectedTrigger, nil)
	//			if tc.waitForBuilds {
	//				mockBuildService.On("WaitForBuild", tc.expectedTrigger.BuildType.Name, tc.expectedTrigger.ID, tc.waitTimeout).Return(tc.expectedWait, tc.waitError)
	//				mockBuildService.On("GetBuildStatus", tc.expectedTrigger.ID).Return(types.BuildStatusResponse{Status: "SUCCESS", State: "finished"}, nil)
	//			}
	//			if param.DownloadArtifacts {
	//				mockArtifactsService.On("BuildHasArtifact", tc.expectedTrigger.ID).Return(true)
	//				mockArtifactsService.On("DownloadAndUnzipArtifacts", tc.expectedTrigger.ID, param.BuildTypeID, tc.multiArtifactsPath).Return(nil)
	//				mockArtifactsService.On("GetArtifactChildren", tc.expectedTrigger.ID).Return(types.ArtifactChildren{}, nil)
	//				mockArtifactsService.On("GetAllBuildTypeArtifacts", tc.expectedTrigger.ID, param.BuildTypeID).Return([]byte{}, nil)
	//			}
	//		}
	//
	//		// Call the function
	//		triggerBuilds(client, tc.parameters, tc.waitForBuilds, tc.waitTimeout, tc.multiArtifactsPath, tc.requireArtifacts)
	//
	//		// Verify all expectations
	//		mockBuildService.AssertExpectations(t)
	//		mockArtifactsService.AssertExpectations(t)
	//	})
	//}
}
