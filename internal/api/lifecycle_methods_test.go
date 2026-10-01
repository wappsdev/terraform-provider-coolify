package api_test

import (
	"net/http"
	"testing"

	"terraform-provider-coolify/internal/api"
)

// Coolify v4.3.23 moved start/stop/restart to POST; GET now returns 405
// ("This endpoint has changed to a POST request.").
func TestLifecycleEndpointsUsePOST(t *testing.T) {
	const server = "https://coolify.example.com/api/v1"
	const uuid = "abc123"

	builders := map[string]func() (*http.Request, error){
		"application start":   func() (*http.Request, error) { return api.NewStartApplicationByUuidRequest(server, uuid, nil) },
		"application stop":    func() (*http.Request, error) { return api.NewStopApplicationByUuidRequest(server, uuid) },
		"application restart": func() (*http.Request, error) { return api.NewRestartApplicationByUuidRequest(server, uuid) },
		"database start":      func() (*http.Request, error) { return api.NewStartDatabaseByUuidRequest(server, uuid) },
		"database stop":       func() (*http.Request, error) { return api.NewStopDatabaseByUuidRequest(server, uuid) },
		"database restart":    func() (*http.Request, error) { return api.NewRestartDatabaseByUuidRequest(server, uuid) },
		"service start":       func() (*http.Request, error) { return api.NewStartServiceByUuidRequest(server, uuid) },
		"service stop":        func() (*http.Request, error) { return api.NewStopServiceByUuidRequest(server, uuid) },
		"service restart":     func() (*http.Request, error) { return api.NewRestartServiceByUuidRequest(server, uuid) },
	}

	for name, build := range builders {
		req, err := build()
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if req.Method != http.MethodPost {
			t.Errorf("%s: want POST, got %s", name, req.Method)
		}
	}
}
