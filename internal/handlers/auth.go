package handlers

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
)

// BasicAuthHandler handles HTTP Basic Authentication
func BasicAuthHandler(w http.ResponseWriter, r *http.Request) {
	// Extract expected credentials from path: /basic-auth/{user}/{passwd}
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/basic-auth/"), "/")
	if len(pathParts) < 2 {
		writeJSONError(w, http.StatusBadRequest, "Invalid path format. Use /basic-auth/{user}/{passwd}")
		return
	}

	expectedUser := pathParts[0]
	expectedPasswd := pathParts[1]

	// Get Authorization header
	auth := r.Header.Get("Authorization")
	if auth == "" {
		// Send 401 with WWW-Authenticate header
		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
		writeJSONError(w, http.StatusUnauthorized, "Authorization required")
		return
	}

	// Check if it's Basic auth
	if !strings.HasPrefix(auth, "Basic ") {
		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
		writeJSONError(w, http.StatusUnauthorized, "Basic authentication required")
		return
	}

	// Decode base64 credentials
	payload, err := base64.StdEncoding.DecodeString(auth[6:])
	if err != nil {
		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
		writeJSONError(w, http.StatusUnauthorized, "Invalid authorization format")
		return
	}

	// Split user:password
	credentials := strings.SplitN(string(payload), ":", 2)
	if len(credentials) != 2 {
		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
		writeJSONError(w, http.StatusUnauthorized, "Invalid credentials format")
		return
	}

	user := credentials[0]
	passwd := credentials[1]

	// Compare credentials
	if user != expectedUser || passwd != expectedPasswd {
		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
		writeJSONError(w, http.StatusUnauthorized, "Invalid username or password")
		return
	}

	// Authentication successful
	response := map[string]any{
		"authenticated": true,
		"user":          user,
	}
	writeJSONResponse(w, http.StatusOK, response)
}

// BearerHandler handles Bearer token authentication
func BearerHandler(w http.ResponseWriter, r *http.Request) {
	// Get Authorization header
	auth := r.Header.Get("Authorization")
	if auth == "" {
		w.Header().Set("WWW-Authenticate", `Bearer realm="Restricted"`)
		writeJSONError(w, http.StatusUnauthorized, "Authorization required")
		return
	}

	// Check if it's Bearer auth
	if !strings.HasPrefix(auth, "Bearer ") {
		w.Header().Set("WWW-Authenticate", `Bearer realm="Restricted"`)
		writeJSONError(w, http.StatusUnauthorized, "Bearer token required")
		return
	}

	// Extract token
	token := strings.TrimPrefix(auth, "Bearer ")
	if token == "" {
		w.Header().Set("WWW-Authenticate", `Bearer realm="Restricted"`)
		writeJSONError(w, http.StatusUnauthorized, "Bearer token is empty")
		return
	}

	// For this simple implementation, any non-empty token is valid
	// In production, you would validate the token against a database or JWT
	response := map[string]any{
		"authenticated": true,
		"token":         token,
	}
	writeJSONResponse(w, http.StatusOK, response)
}

// DigestAuthHandler handles HTTP Digest Authentication
func DigestAuthHandler(w http.ResponseWriter, r *http.Request) {
	// Extract parameters from path: /digest-auth/{qop}/{user}/{passwd}
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/digest-auth/"), "/")
	if len(pathParts) < 3 {
		writeJSONError(w, http.StatusBadRequest, "Invalid path format. Use /digest-auth/{qop}/{user}/{passwd}")
		return
	}

	qop := pathParts[0]
	expectedUser := pathParts[1]
	_ = pathParts[2] // expectedPasswd - not validated in simplified implementation

	// Get Authorization header
	auth := r.Header.Get("Authorization")
	if auth == "" {
		// Send 401 with WWW-Authenticate header for digest auth
		nonce := generateNonce()
		opaque := generateOpaque()
		challenge := fmt.Sprintf(`Digest realm="Restricted", qop="%s", nonce="%s", opaque="%s"`, qop, nonce, opaque)
		w.Header().Set("WWW-Authenticate", challenge)
		writeJSONError(w, http.StatusUnauthorized, "Authorization required")
		return
	}

	// Check if it's Digest auth
	if !strings.HasPrefix(auth, "Digest ") {
		nonce := generateNonce()
		opaque := generateOpaque()
		challenge := fmt.Sprintf(`Digest realm="Restricted", qop="%s", nonce="%s", opaque="%s"`, qop, nonce, opaque)
		w.Header().Set("WWW-Authenticate", challenge)
		writeJSONError(w, http.StatusUnauthorized, "Digest authentication required")
		return
	}

	// Parse digest auth parameters
	digestParams := parseDigestAuth(auth[7:])

	// Validate username
	username, ok := digestParams["username"]
	if !ok || username != expectedUser {
		nonce := generateNonce()
		opaque := generateOpaque()
		challenge := fmt.Sprintf(`Digest realm="Restricted", qop="%s", nonce="%s", opaque="%s"`, qop, nonce, opaque)
		w.Header().Set("WWW-Authenticate", challenge)
		writeJSONError(w, http.StatusUnauthorized, "Invalid username")
		return
	}

	// For simplified implementation, we'll accept valid format with correct username
	// Full RFC 2617 implementation would require validating the response hash
	response := map[string]any{
		"authenticated": true,
		"user":          username,
	}
	writeJSONResponse(w, http.StatusOK, response)
}

// generateNonce generates a random nonce for digest auth
func generateNonce() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// generateOpaque generates a random opaque value for digest auth
func generateOpaque() string {
	b := make([]byte, 16)
	rand.Read(b)
	hash := md5.Sum(b)
	return hex.EncodeToString(hash[:])
}

// parseDigestAuth parses digest authentication parameters
func parseDigestAuth(auth string) map[string]string {
	params := make(map[string]string)
	parts := strings.Split(auth, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}

		key := strings.TrimSpace(kv[0])
		value := strings.Trim(strings.TrimSpace(kv[1]), `"`)
		params[key] = value
	}

	return params
}
