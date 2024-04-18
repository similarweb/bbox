package teamcity

//
//import (
//	"net/http"
//	"net/http/httptest"
//	"testing"
//
//	"github.com/stretchr/testify/assert"
//)
//
//func TestClearQueue(t *testing.T) {
//	// Create a mock HTTP server
//	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		// Check that the method is DELETE
//		if r.Method != http.MethodDelete {
//			t.Errorf("Expected 'DELETE' request, got '%s'", r.Method)
//		}
//
//		// Respond with a 204 No Content, which is a typical success status for this API
//		w.WriteHeader(http.StatusNoContent)
//	}))
//	defer server.Close()
//
//	// Assume setupTeamCityClient initializes a client that includes setting the server URL to the mock server's URL.
//	// You will need to modify your client struct and methods accordingly to allow injecting the base URL (i.e., server.URL).
//	client, teardown := setupTeamCityClient(server.URL) // setupTeamCityClient needs to be implemented
//	defer teardown()
//
//	// Perform the test
//	err := client.Queue.ClearQueue()
//	assert.NoError(t, err, "ClearQueue should not return an error with HTTP 204")
//}
//
//// setupTeamCityClient must be adapted to accept a URL and use it as the base URL for the HTTP client in your actual implementation
//func setupTeamCityClient(url string) (*Client, func()) {
//	client := Client{
//		client:  &http.Client{},
//		Queue:   (*QueueService)(&service{&Client{}}),
//		baseURL: url,
//	}
//	return &client, func() {}
//}
