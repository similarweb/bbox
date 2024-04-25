package multitrigger

import (
	"testing"

	"bbox/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		{
			name: "invalid combination format",
			combinations: []string{
				"projectID;main;true",
			},
			expectedError: "invalid combination format",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			output, err := parseCombinations(tc.combinations)

			if tc.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedOutput, output)
			}
		})
	}
}
