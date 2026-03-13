package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestLoggingMiddleware tests that the logging middleware doesn't break the request flow
func TestLoggingMiddleware(t *testing.T) {
	// Create a test handler that returns 200
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})

	// Wrap with logging middleware
	wrapped := Logging(handler)

	// Create test request
	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	// Execute request
	wrapped.ServeHTTP(rr, req)

	// Check response
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	if rr.Body.String() != "test response" {
		t.Errorf("Expected body 'test response', got '%s'", rr.Body.String())
	}
}

// TestLoggingMiddlewareWithError tests logging middleware with error responses
func TestLoggingMiddlewareWithError(t *testing.T) {
	// Create a test handler that returns 500
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error"))
	})

	// Wrap with logging middleware
	wrapped := Logging(handler)

	// Create test request
	req := httptest.NewRequest("POST", "/error", nil)
	rr := httptest.NewRecorder()

	// Execute request
	wrapped.ServeHTTP(rr, req)

	// Check response
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", rr.Code)
	}

	if rr.Body.String() != "error" {
		t.Errorf("Expected body 'error', got '%s'", rr.Body.String())
	}
}

// TestLoggingMiddlewareMultipleWrites tests that status code is captured correctly
func TestLoggingMiddlewareMultipleWrites(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("part1"))
		w.Write([]byte("part2"))
	})

	wrapped := Logging(handler)
	req := httptest.NewRequest("POST", "/test", nil)
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", rr.Code)
	}

	if rr.Body.String() != "part1part2" {
		t.Errorf("Expected body 'part1part2', got '%s'", rr.Body.String())
	}
}

// TestResponseWriterStatusCode tests that responseWriter captures status codes correctly
func TestResponseWriterStatusCode(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		expectedStatus int
	}{
		{"200 OK", http.StatusOK, http.StatusOK},
		{"404 Not Found", http.StatusNotFound, http.StatusNotFound},
		{"500 Internal Server Error", http.StatusInternalServerError, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			})

			wrapped := Logging(handler)
			req := httptest.NewRequest("GET", "/test", nil)
			rr := httptest.NewRecorder()

			wrapped.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}

// TestResponseWriterImplicitOK tests that responseWriter sets 200 OK when WriteHeader is not called
func TestResponseWriterImplicitOK(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Don't call WriteHeader explicitly
		w.Write([]byte("test"))
	})

	wrapped := Logging(handler)
	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200 (implicit), got %d", rr.Code)
	}
}
