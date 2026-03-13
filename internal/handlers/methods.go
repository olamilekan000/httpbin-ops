package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// RequestInfo represents the details of an HTTP request
type RequestInfo struct {
	Method  string              `json:"method"`
	URL     string              `json:"url"`
	Args    map[string][]string `json:"args"`
	Headers map[string][]string `json:"headers"`
	Origin  string              `json:"origin"`
	Body    string              `json:"body,omitempty"`
	JSON    any                 `json:"json,omitempty"`
}

// extractRequestInfo extracts information from an HTTP request
func extractRequestInfo(r *http.Request) (*RequestInfo, error) {
	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	info := &RequestInfo{
		Method:  r.Method,
		URL:     r.URL.String(),
		Args:    r.URL.Query(),
		Headers: r.Header,
		Origin:  getOriginIP(r),
		Body:    string(body),
	}

	// Try to parse JSON body if Content-Type is application/json
	if len(body) > 0 && strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		var jsonData any
		if err := json.Unmarshal(body, &jsonData); err == nil {
			info.JSON = jsonData
		}
	}

	return info, nil
}

// getOriginIP extracts the origin IP from the request
func getOriginIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	// Remove port if present
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}

// writeJSONResponse writes a JSON response
func writeJSONResponse(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeJSONError writes a JSON error response
func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSONResponse(w, status, map[string]string{"error": message})
}

// MethodHandler returns a handler for a specific HTTP method
func MethodHandler(method string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check if the request method matches
		if r.Method != method {
			writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		// Extract request information
		info, err := extractRequestInfo(r)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "Failed to read request body")
			return
		}

		// For HEAD requests, only send headers, no body
		if method == "HEAD" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			return
		}

		// For OPTIONS, include Allow header
		if method == "OPTIONS" {
			w.Header().Set("Allow", "GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS")
		}

		writeJSONResponse(w, http.StatusOK, info)
	}
}

// HeadersHandler returns all request headers
func HeadersHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]any{
		"headers": r.Header,
	}
	writeJSONResponse(w, http.StatusOK, response)
}

// IPHandler returns the origin IP address
func IPHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"origin": getOriginIP(r),
	}
	writeJSONResponse(w, http.StatusOK, response)
}

// UserAgentHandler returns the User-Agent header
func UserAgentHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"user-agent": r.Header.Get("User-Agent"),
	}
	writeJSONResponse(w, http.StatusOK, response)
}

// DelayHandler delays the response for a specified number of seconds
func DelayHandler(w http.ResponseWriter, r *http.Request) {
	// Extract delay from path: /delay/{seconds}
	path := strings.TrimPrefix(r.URL.Path, "/delay/")
	seconds, err := strconv.Atoi(path)
	if err != nil || seconds < 0 {
		writeJSONError(w, http.StatusBadRequest, "Invalid delay value")
		return
	}

	// Cap delay at 10 seconds
	if seconds > 10 {
		seconds = 10
	}

	// Sleep for the specified duration
	time.Sleep(time.Duration(seconds) * time.Second)

	// Extract and return request info
	info, err := extractRequestInfo(r)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to read request body")
		return
	}

	writeJSONResponse(w, http.StatusOK, info)
}
