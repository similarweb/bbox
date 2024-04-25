package teamcity

//
//import (
//	"bbox/pkg/types"
//	"bbox/pkg/utils/testutils"
//	"github.com/stretchr/testify/assert"
//	"net/http"
//	"testing"
//)
//
//func TestTriggerBuild(t *testing.T) {
//	config := testutils.MockServerConfig{
//		Endpoints: []testutils.EndpointConfig{
//			{
//				Path:       "/httpAuth/app/rest/buildQueue",
//				Method:     "POST",
//				Response:   types.TriggerBuildWithParametersResponse{ID: 123, State: "SUCCESS"},
//				StatusCode: http.StatusOK,
//			},
//		},
//	}
//
//	server, baseURL := testutils.SetupMockServer(config)
//	defer server.Close()
//
//	client := NewTeamCityClient(baseURL, "testuser", "testpassword")
//
//	// Test the TriggerBuild function
//	triggerResponse, err := client.Build.TriggerBuild("typeID", "branchName", map[string]string{"param1": "value1"})
//	assert.NoError(t, err)
//	assert.Equal(t, "SUCCESS", triggerResponse.State)
//}
//
//func TestGetBuildStatus(t *testing.T) {
//	config := testutils.MockServerConfig{
//		Endpoints: []testutils.EndpointConfig{
//			{
//				Path:       "/app/rest/builds/id:123",
//				Method:     "GET",
//				Response:   types.BuildStatusResponse{ID: 123, State: "SUCCESS", Status: "Completed"},
//				StatusCode: http.StatusOK,
//			},
//		},
//	}
//
//	server, baseURL := testutils.SetupMockServer(config)
//	defer server.Close()
//
//	client := NewTeamCityClient(baseURL, "testuser", "testpassword")
//
//	// Test the GetBuildStatus function
//	buildStatus, err := client.Build.GetBuildStatus(123)
//	assert.NoError(t, err)
//	assert.Equal(t, 123, buildStatus.ID)
//	assert.Equal(t, "SUCCESS", buildStatus.State)
//	assert.Equal(t, "Completed", buildStatus.Status)
//}

//package teamcity
//
//import (
//	"bbox/pkg/utils/testutils"
//	"bytes"
//	"io/ioutil"
//	"net/http"
//	"net/url"
//	"testing"
//
//	"github.com/stretchr/testify/assert"
//	"github.com/stretchr/testify/mock"
//)
//
//func TestGetBuildStatus(t *testing.T) {
//	// Create an instance of the mock HTTP client
//	mockHttp := new(testutils.MockHttpClient)
//	client := &Client{
//		baseURL:   &url.URL{Scheme: "http", Host: "example.com"},
//		client:    mockHttp,
//		BasicAuth: &BasicAuth{username: "user", password: "pass"},
//	}
//
//	bs := &BuildService{client: client}
//
//	// Setup the mock expectations
//	mockResp := &http.Response{
//		StatusCode: http.StatusOK,
//		Body:       ioutil.NopCloser(bytes.NewBufferString(`{"status": "SUCCESS", "state": "finished"}`)),
//	}
//	mockHttp.On("Do", mock.AnythingOfType("*http.Request")).Return(mockResp, nil)
//
//	// Call the method
//	result, err := bs.GetBuildStatus(123)
//
//	// Assertions
//	assert.NoError(t, err)
//	assert.NotNil(t, result)
//	assert.Equal(t, "SUCCESS", result.Status)
//	mockHttp.AssertExpectations(t) // Verify that the expectations were met
//}
