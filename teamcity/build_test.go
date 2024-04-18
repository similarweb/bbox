package teamcity

import (
	"bbox/pkg/types"
	"bbox/pkg/utils/testutils"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestTriggerBuild(t *testing.T) {
	config := testutils.MockServerConfig{
		Endpoints: []testutils.EndpointConfig{
			{
				Path:       "/httpAuth/app/rest/buildQueue",
				Method:     "POST",
				Response:   types.TriggerBuildWithParametersResponse{ID: 123, State: "SUCCESS"},
				StatusCode: http.StatusOK,
			},
		},
	}

	server, baseURL := testutils.SetupMockServer(config)
	defer server.Close()

	client := NewTeamCityClient(baseURL, "testuser", "testpassword")

	// Test the TriggerBuild function
	triggerResponse, err := client.Build.TriggerBuild("typeID", "branchName", map[string]string{"param1": "value1"})
	assert.NoError(t, err)
	assert.Equal(t, "SUCCESS", triggerResponse.State)
}

func TestGetBuildStatus(t *testing.T) {
	config := testutils.MockServerConfig{
		Endpoints: []testutils.EndpointConfig{
			{
				Path:       "/app/rest/builds/id:123",
				Method:     "GET",
				Response:   types.BuildStatusResponse{ID: 123, State: "SUCCESS", Status: "Completed"},
				StatusCode: http.StatusOK,
			},
		},
	}

	server, baseURL := testutils.SetupMockServer(config)
	defer server.Close()

	client := NewTeamCityClient(baseURL, "testuser", "testpassword")

	// Test the GetBuildStatus function
	buildStatus, err := client.Build.GetBuildStatus(123)
	assert.NoError(t, err)
	assert.Equal(t, 123, buildStatus.ID)
	assert.Equal(t, "SUCCESS", buildStatus.State)
	assert.Equal(t, "Completed", buildStatus.Status)
}
