package server

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestServerRouting tests that all routes are properly configured
func TestServerRouting(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		authHeader     string
		expectedStatus int
	}{
		{"GET endpoint", "GET", "/get", "", http.StatusOK},
		{"POST endpoint", "POST", "/post", "", http.StatusOK},
		{"PUT endpoint", "PUT", "/put", "", http.StatusOK},
		{"PATCH endpoint", "PATCH", "/patch", "", http.StatusOK},
		{"DELETE endpoint", "DELETE", "/delete", "", http.StatusOK},
		{"HEAD endpoint", "HEAD", "/head", "", http.StatusOK},
		{"OPTIONS endpoint", "OPTIONS", "/options", "", http.StatusOK},
		{"Headers endpoint", "GET", "/headers", "", http.StatusOK},
		{"IP endpoint", "GET", "/ip", "", http.StatusOK},
		{"User-Agent endpoint", "GET", "/user-agent", "", http.StatusOK},
		{"Status endpoint", "GET", "/status/200", "", http.StatusOK},
		{"Status 404", "GET", "/status/404", "", http.StatusNotFound},
		{"Delay endpoint", "GET", "/delay/0", "", http.StatusOK},
		{"Basic Auth - no auth", "GET", "/basic-auth/user/passwd", "", http.StatusUnauthorized},
		{"Basic Auth - valid", "GET", "/basic-auth/user/passwd", "Basic " + base64.StdEncoding.EncodeToString([]byte("user:passwd")), http.StatusOK},
		{"Bearer - no auth", "GET", "/bearer", "", http.StatusUnauthorized},
		{"Bearer - valid", "GET", "/bearer", "Bearer token123", http.StatusOK},
		{"Digest Auth - no auth", "GET", "/digest-auth/auth/user/passwd", "", http.StatusUnauthorized},
	}

	// Create server
	srv := New(":8080")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			rr := httptest.NewRecorder()
			srv.mux.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}

// TestServerIntegration tests the full server integration
func TestServerIntegration(t *testing.T) {
	// Create a test server
	srv := New(":0")
	testServer := httptest.NewServer(srv.mux)
	defer testServer.Close()

	t.Run("GET with query params", func(t *testing.T) {
		resp, err := http.Get(testServer.URL + "/get?foo=bar&baz=qux")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		body, _ := io.ReadAll(resp.Body)
		var data map[string]any
		if err := json.Unmarshal(body, &data); err != nil {
			t.Errorf("Failed to parse JSON: %v", err)
		}

		if data["method"] != "GET" {
			t.Errorf("Expected method GET, got %v", data["method"])
		}

		args, ok := data["args"].(map[string]any)
		if !ok || len(args) == 0 {
			t.Error("Expected query args to be present")
		}
	})

	t.Run("POST with JSON body", func(t *testing.T) {
		jsonBody := `{"key":"value","number":42}`
		resp, err := http.Post(testServer.URL+"/post", "application/json", strings.NewReader(jsonBody))
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		body, _ := io.ReadAll(resp.Body)
		var data map[string]any
		if err := json.Unmarshal(body, &data); err != nil {
			t.Errorf("Failed to parse JSON: %v", err)
		}

		if data["method"] != "POST" {
			t.Errorf("Expected method POST, got %v", data["method"])
		}

		if data["json"] == nil {
			t.Error("Expected JSON body to be parsed")
		}
	})

	t.Run("Status code endpoint", func(t *testing.T) {
		resp, err := http.Get(testServer.URL + "/status/418")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 418 {
			t.Errorf("Expected status 418, got %d", resp.StatusCode)
		}
	})

	t.Run("Headers endpoint", func(t *testing.T) {
		req, _ := http.NewRequest("GET", testServer.URL+"/headers", nil)
		req.Header.Set("X-Custom-Header", "custom-value")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		body, _ := io.ReadAll(resp.Body)
		var data map[string]any
		if err := json.Unmarshal(body, &data); err != nil {
			t.Errorf("Failed to parse JSON: %v", err)
		}

		if data["headers"] == nil {
			t.Error("Expected headers to be present")
		}
	})

	t.Run("Basic Auth success", func(t *testing.T) {
		req, _ := http.NewRequest("GET", testServer.URL+"/basic-auth/testuser/testpass", nil)
		req.SetBasicAuth("testuser", "testpass")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		body, _ := io.ReadAll(resp.Body)
		var data map[string]any
		if err := json.Unmarshal(body, &data); err != nil {
			t.Errorf("Failed to parse JSON: %v", err)
		}

		if authenticated, ok := data["authenticated"].(bool); !ok || !authenticated {
			t.Error("Expected authenticated to be true")
		}

		if user, ok := data["user"].(string); !ok || user != "testuser" {
			t.Errorf("Expected user to be 'testuser', got '%s'", user)
		}
	})

	t.Run("Basic Auth failure", func(t *testing.T) {
		req, _ := http.NewRequest("GET", testServer.URL+"/basic-auth/testuser/testpass", nil)
		req.SetBasicAuth("testuser", "wrongpass")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", resp.StatusCode)
		}

		if auth := resp.Header.Get("WWW-Authenticate"); auth == "" {
			t.Error("Expected WWW-Authenticate header")
		}
	})

	t.Run("Bearer token", func(t *testing.T) {
		req, _ := http.NewRequest("GET", testServer.URL+"/bearer", nil)
		req.Header.Set("Authorization", "Bearer my-secret-token")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		body, _ := io.ReadAll(resp.Body)
		var data map[string]any
		if err := json.Unmarshal(body, &data); err != nil {
			t.Errorf("Failed to parse JSON: %v", err)
		}

		if token, ok := data["token"].(string); !ok || token != "my-secret-token" {
			t.Errorf("Expected token to be 'my-secret-token', got '%s'", token)
		}
	})
}

// TestServerGracefulShutdown tests that the server can shut down gracefully
func TestServerGracefulShutdown(t *testing.T) {
	srv := New(":0")
	testServer := httptest.NewServer(srv.mux)

	// Make a request to ensure server is running
	resp, err := http.Get(testServer.URL + "/get")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Close the server
	testServer.Close()

	// Verify server is stopped by attempting another request
	_, err = http.Get(testServer.URL + "/get")
	if err == nil {
		t.Error("Expected error when connecting to closed server")
	}
}
