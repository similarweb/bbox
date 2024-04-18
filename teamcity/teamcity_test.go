package teamcity

//
//import (
//	"net/url"
//	"testing"
//
//	"github.com/stretchr/testify/assert"
//)
//
//func TestNewRequestWrapper(t *testing.T) {
//	baseURL, _ := url.Parse("https://example.com/")
//
//	client := NewTeamCityClient(baseURL, "user", "pass")
//
//	tests := []struct {
//		name    string
//		method  string
//		urlStr  string
//		body    interface{}
//		wantErr bool
//	}{
//		{
//			name:    "Valid request with nil body",
//			method:  "GET",
//			urlStr:  "path/",
//			body:    nil,
//			wantErr: false,
//		},
//		{
//			name:    "Valid request with body",
//			method:  "POST",
//			urlStr:  "path/",
//			body:    map[string]interface{}{"key": "value"},
//			wantErr: false,
//		},
//		{
//			name:    "Invalid URL",
//			method:  "GET",
//			urlStr:  ":",
//			wantErr: true,
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			req, err := client.NewRequestWrapper(tt.method, tt.urlStr, tt.body)
//			if (err != nil) != tt.wantErr {
//				t.Errorf("NewRequestWrapper() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//
//			if err == nil {
//				assert.Equal(t, tt.method, req.Method, "HTTP method mismatch")
//				parsedURL, _ := url.Parse(tt.urlStr)
//				assert.Equal(t, client.baseURL.ResolveReference(parsedURL).String(), req.URL.String(), "URL mismatch")
//
//				if tt.body != nil {
//					contentType := req.Header.Get("Content-Type")
//					assert.Equal(t, "application/json", contentType, "Content-Type header should be application/json")
//				}
//
//				username, password, _ := req.BasicAuth()
//				assert.Equal(t, "user", username, "Basic auth username mismatch")
//				assert.Equal(t, "pass", password, "Basic auth password mismatch")
//			}
//		})
//	}
//}
