package teamcity

import (
	"bbox/pkg/types"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildService_WaitForBuild(t *testing.T) {
	tests := []struct {
		name           string
		mockResponse   types.BuildStatusResponse // Single response that represents finished state
		expectedResult types.BuildStatusResponse
		expectedError  string
	}{
		{
			name:           "Successful build",
			mockResponse:   types.BuildStatusResponse{ID: 123, Status: "SUCCESS", State: "finished"},
			expectedResult: types.BuildStatusResponse{ID: 123, Status: "SUCCESS", State: "finished"},
			expectedError:  "",
		},
		{
			name:           "Failed build",
			mockResponse:   types.BuildStatusResponse{ID: 123, Status: "FAILURE", State: "finished"},
			expectedResult: types.BuildStatusResponse{ID: 123, Status: "FAILURE", State: "finished"},
			expectedError:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test server that returns finished state immediately
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			// Parse server URL
			serverURL, err := url.Parse(server.URL)
			require.NoError(t, err)

			// Create client
			client, err := NewTeamCityClient(serverURL, "testuser", "testpass")
			require.NoError(t, err)

			// Call WaitForBuild
			result, err := client.Build.WaitForBuild("TestBuild", 123, 30*time.Second)

			// Assert results
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedResult.ID, result.ID)
			assert.Equal(t, tt.expectedResult.Status, result.Status)
			assert.Equal(t, tt.expectedResult.State, result.State)
		})
	}
}
