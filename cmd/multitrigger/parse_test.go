package multitrigger

import (
	"bbox/pkg/types"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseCombinations(t *testing.T) {
	testCases := []struct {
		name           string
		combinations   []string
		expectedOutput []types.BuildParameters
		expectedError  string
	}{
		{
			name: "valid combinations",
			combinations: []string{
				"bt1;main;true;key1=value1&key2=value2",
			},
			expectedOutput: []types.BuildParameters{
				{
					BuildTypeID:       "bt1",
					BranchName:        "main",
					DownloadArtifacts: true,
					PropertiesFlag: map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
				},
			},
		},
		{
			name: "valid combinations no artifacts",
			combinations: []string{
				"bt1;main;0;key1=value1&key2=value2",
			},
			expectedOutput: []types.BuildParameters{
				{
					BuildTypeID:       "bt1",
					BranchName:        "main",
					DownloadArtifacts: false,
					PropertiesFlag: map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
				},
			},
		},
		{
			name: "invalid combination format",
			combinations: []string{
				"invalid_format",
			},
			expectedError: "invalid combination format: invalid_format",
		},
		{
			name: "invalid downloadArtifacts boolean",
			combinations: []string{
				"projectID;main;tru;key1=value1&key2=value2",
			},
			expectedError: "invalid downloadArtifacts boolean",
		},
		{
			name: "invalid buildID",
			combinations: []string{
				"NOT/AN/ID;main;tru;key1=value1&key2=value2",
			},
			expectedError: "invalid buildTypeID",
		},
		{
			name: "invalid branchName",
			combinations: []string{
				"projectID;~bad_branch;true;key1=value1&key2=value2",
			},
			expectedError: "invalid branchName",
		},
		{
			name: "empty parameter key",
			combinations: []string{
				"projectID;main;true;=value1&key2=value2",
			},
			expectedError: "invalid property key",
		},
		{
			name: "invalid parameter format",
			combinations: []string{
				"projectID;main;true;m&key1=value1&key2=value2",
			},
			expectedError: "invalid property format",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output, err := parseCombinations(tc.combinations)

			if tc.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedOutput, output)
			}
		})
	}
}
