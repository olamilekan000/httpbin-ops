package handlers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestMethodHandler tests HTTP method handlers
func TestMethodHandler(t *testing.T) {
	tests := []struct {
		name           string
		handlerMethod  string
		requestMethod  string
		path           string
		body           string
		contentType    string
		expectedStatus int
		checkResponse  bool
	}{
		{
			name:           "GET request",
			handlerMethod:  "GET",
			requestMethod:  "GET",
			path:           "/get?foo=bar",
			expectedStatus: http.StatusOK,
			checkResponse:  true,
		},
		{
			name:           "POST request with JSON",
			handlerMethod:  "POST",
			requestMethod:  "POST",
			path:           "/post",
			body:           `{"key":"value"}`,
			contentType:    "application/json",
			expectedStatus: http.StatusOK,
			checkResponse:  true,
		},
		{
			name:           "PUT request",
			handlerMethod:  "PUT",
			requestMethod:  "PUT",
			path:           "/put",
			body:           `test body`,
			expectedStatus: http.StatusOK,
			checkResponse:  true,
		},
		{
			name:           "PATCH request",
			handlerMethod:  "PATCH",
			requestMethod:  "PATCH",
			path:           "/patch",
			expectedStatus: http.StatusOK,
			checkResponse:  true,
		},
		{
			name:           "DELETE request",
			handlerMethod:  "DELETE",
			requestMethod:  "DELETE",
			path:           "/delete",
			expectedStatus: http.StatusOK,
			checkResponse:  true,
		},
		{
			name:           "HEAD request",
			handlerMethod:  "HEAD",
			requestMethod:  "HEAD",
			path:           "/head",
			expectedStatus: http.StatusOK,
			checkResponse:  false, // HEAD should not return body
		},
		{
			name:           "OPTIONS request",
			handlerMethod:  "OPTIONS",
			requestMethod:  "OPTIONS",
			path:           "/options",
			expectedStatus: http.StatusOK,
			checkResponse:  true,
		},
		{
			name:           "Wrong method",
			handlerMethod:  "GET",
			requestMethod:  "POST",
			path:           "/get",
			expectedStatus: http.StatusMethodNotAllowed,
			checkResponse:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.body != "" {
				req = httptest.NewRequest(tt.requestMethod, tt.path, strings.NewReader(tt.body))
			} else {
				req = httptest.NewRequest(tt.requestMethod, tt.path, nil)
			}

			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			rr := httptest.NewRecorder()
			handler := MethodHandler(tt.handlerMethod)
			handler(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.checkResponse && rr.Code == http.StatusOK {
				var response RequestInfo
				if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
					t.Errorf("Failed to decode response: %v", err)
				}

				if response.Method != tt.requestMethod {
					t.Errorf("Expected method %s, got %s", tt.requestMethod, response.Method)
				}

				if tt.requestMethod == "OPTIONS" {
					allowHeader := rr.Header().Get("Allow")
					if allowHeader == "" {
						t.Error("Expected Allow header in OPTIONS response")
					}
				}
			}
		})
	}
}

// TestHeadersHandler tests the headers endpoint
func TestHeadersHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/headers", nil)
	req.Header.Set("User-Agent", "test-agent")
	req.Header.Set("X-Custom", "custom-value")

	rr := httptest.NewRecorder()
	HeadersHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	var response map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	headers, ok := response["headers"].(map[string]any)
	if !ok {
		t.Error("Response does not contain headers")
	}

	if headers == nil {
		t.Error("Headers map is nil")
	}
}

// TestIPHandler tests the IP endpoint
func TestIPHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/ip", nil)
	req.RemoteAddr = "192.168.1.1:12345"

	rr := httptest.NewRecorder()
	IPHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if origin, ok := response["origin"]; !ok || origin == "" {
		t.Error("Response does not contain origin IP")
	}
}

// TestUserAgentHandler tests the user-agent endpoint
func TestUserAgentHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/user-agent", nil)
	req.Header.Set("User-Agent", "test-agent/1.0")

	rr := httptest.NewRecorder()
	UserAgentHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if ua := response["user-agent"]; ua != "test-agent/1.0" {
		t.Errorf("Expected user-agent 'test-agent/1.0', got '%s'", ua)
	}
}

// TestDelayHandler tests the delay endpoint
func TestDelayHandler(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		expectedStatus int
		minDuration    time.Duration
		maxDuration    time.Duration
	}{
		{
			name:           "1 second delay",
			path:           "/delay/1",
			expectedStatus: http.StatusOK,
			minDuration:    900 * time.Millisecond,
			maxDuration:    1200 * time.Millisecond,
		},
		{
			name:           "0 second delay",
			path:           "/delay/0",
			expectedStatus: http.StatusOK,
			minDuration:    0,
			maxDuration:    100 * time.Millisecond,
		},
		{
			name:           "Invalid delay",
			path:           "/delay/abc",
			expectedStatus: http.StatusBadRequest,
			minDuration:    0,
			maxDuration:    100 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			rr := httptest.NewRecorder()

			start := time.Now()
			DelayHandler(rr, req)
			duration := time.Since(start)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if duration < tt.minDuration || duration > tt.maxDuration {
				t.Errorf("Expected duration between %v and %v, got %v", tt.minDuration, tt.maxDuration, duration)
			}
		})
	}
}

// TestStatusHandler tests the status code endpoint
func TestStatusHandler(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		expectedStatus int
	}{
		{
			name:           "200 OK",
			path:           "/status/200",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "404 Not Found",
			path:           "/status/404",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "500 Internal Server Error",
			path:           "/status/500",
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Invalid status code",
			path:           "/status/abc",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Status code out of range",
			path:           "/status/999",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			rr := httptest.NewRecorder()

			StatusHandler(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}

// TestStatusHandlerWeighted tests weighted random status codes
func TestStatusHandlerWeighted(t *testing.T) {
	req := httptest.NewRequest("GET", "/status/200:1,404:0", nil)
	rr := httptest.NewRecorder()

	StatusHandler(rr, req)

	// With weight 1 for 200 and 0 for 404, should always return 200
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200 with weight 1, got %d", rr.Code)
	}
}

// TestBasicAuthHandler tests basic authentication
func TestBasicAuthHandler(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		authHeader     string
		expectedStatus int
	}{
		{
			name:           "Valid credentials",
			path:           "/basic-auth/user/passwd",
			authHeader:     "Basic " + base64.StdEncoding.EncodeToString([]byte("user:passwd")),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid credentials",
			path:           "/basic-auth/user/passwd",
			authHeader:     "Basic " + base64.StdEncoding.EncodeToString([]byte("user:wrong")),
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "No auth header",
			path:           "/basic-auth/user/passwd",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Invalid auth format",
			path:           "/basic-auth/user/passwd",
			authHeader:     "Bearer token",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Invalid base64",
			path:           "/basic-auth/user/passwd",
			authHeader:     "Basic invalid!!!",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			rr := httptest.NewRecorder()
			BasicAuthHandler(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if rr.Code == http.StatusUnauthorized {
				if auth := rr.Header().Get("WWW-Authenticate"); auth == "" {
					t.Error("Expected WWW-Authenticate header on 401 response")
				}
			}

			if rr.Code == http.StatusOK {
				var response map[string]any
				if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
					t.Errorf("Failed to decode response: %v", err)
				}

				if authenticated, ok := response["authenticated"].(bool); !ok || !authenticated {
					t.Error("Expected authenticated to be true")
				}
			}
		})
	}
}

// TestBearerHandler tests bearer token authentication
func TestBearerHandler(t *testing.T) {
	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
	}{
		{
			name:           "Valid bearer token",
			authHeader:     "Bearer my-token-123",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Empty bearer token",
			authHeader:     "Bearer ",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "No auth header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Wrong auth type",
			authHeader:     "Basic token",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/bearer", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			rr := httptest.NewRecorder()
			BearerHandler(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if rr.Code == http.StatusUnauthorized {
				if auth := rr.Header().Get("WWW-Authenticate"); auth == "" {
					t.Error("Expected WWW-Authenticate header on 401 response")
				}
			}
		})
	}
}

// TestDigestAuthHandler tests digest authentication
func TestDigestAuthHandler(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		authHeader     string
		expectedStatus int
	}{
		{
			name:           "No auth header",
			path:           "/digest-auth/auth/user/passwd",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Valid digest auth",
			path:           "/digest-auth/auth/user/passwd",
			authHeader:     `Digest username="user", realm="Restricted", nonce="abc123", uri="/digest-auth/auth/user/passwd", response="6629fae49393a05397450978507c4ef1"`,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Wrong username",
			path:           "/digest-auth/auth/user/passwd",
			authHeader:     `Digest username="wrong", realm="Restricted", nonce="abc123", uri="/digest-auth/auth/user/passwd", response="6629fae49393a05397450978507c4ef1"`,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Invalid path format",
			path:           "/digest-auth/auth",
			authHeader:     "",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			rr := httptest.NewRecorder()
			DigestAuthHandler(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if rr.Code == http.StatusUnauthorized {
				if auth := rr.Header().Get("WWW-Authenticate"); auth == "" {
					t.Error("Expected WWW-Authenticate header on 401 response")
				} else if !strings.HasPrefix(auth, "Digest") {
					t.Error("Expected Digest authentication challenge")
				}
			}
		})
	}
}

// TestGetOriginIP tests the getOriginIP function
func TestGetOriginIP(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		xff        string
		xri        string
		expectedIP string
	}{
		{
			name:       "X-Forwarded-For header",
			remoteAddr: "10.0.0.1:12345",
			xff:        "192.168.1.1, 10.0.0.1",
			expectedIP: "192.168.1.1",
		},
		{
			name:       "X-Real-IP header",
			remoteAddr: "10.0.0.1:12345",
			xri:        "192.168.1.1",
			expectedIP: "192.168.1.1",
		},
		{
			name:       "RemoteAddr only",
			remoteAddr: "192.168.1.1:12345",
			expectedIP: "192.168.1.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.xff != "" {
				req.Header.Set("X-Forwarded-For", tt.xff)
			}
			if tt.xri != "" {
				req.Header.Set("X-Real-IP", tt.xri)
			}

			ip := getOriginIP(req)
			if ip != tt.expectedIP {
				t.Errorf("Expected IP %s, got %s", tt.expectedIP, ip)
			}
		})
	}
}

// TestExtractRequestInfo tests the extractRequestInfo function
func TestExtractRequestInfo(t *testing.T) {
	body := `{"test":"data"}`
	req := httptest.NewRequest("POST", "/test?foo=bar&baz=qux", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "test-agent")

	info, err := extractRequestInfo(req)
	if err != nil {
		t.Fatalf("Failed to extract request info: %v", err)
	}

	if info.Method != "POST" {
		t.Errorf("Expected method POST, got %s", info.Method)
	}

	if info.Body != body {
		t.Errorf("Expected body %s, got %s", body, info.Body)
	}

	if info.JSON == nil {
		t.Error("Expected JSON to be parsed")
	}

	if len(info.Args) == 0 {
		t.Error("Expected query args to be parsed")
	}

	if len(info.Headers) == 0 {
		t.Error("Expected headers to be captured")
	}
}
