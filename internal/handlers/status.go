package handlers

import (
	"crypto/rand"
	"encoding/binary"
	"net/http"
	"strconv"
	"strings"
)

// statusWeight represents a status code with its weight
type statusWeight struct {
	code   int
	weight float64
}

// parseStatusCodes parses status code specification from path
// Supports:
//   - Single code: "404"
//   - Weighted codes: "200:0.9,500:0.1"
func parseStatusCodes(path string) ([]statusWeight, error) {
	path = strings.TrimPrefix(path, "/status/")
	if path == "" {
		return nil, nil
	}

	// Check if it contains weights (has colon)
	if !strings.Contains(path, ":") {
		// Simple single status code
		code, err := strconv.Atoi(path)
		if err != nil {
			return nil, err
		}
		return []statusWeight{{code: code, weight: 1.0}}, nil
	}

	// Parse weighted status codes
	var weights []statusWeight
	parts := strings.Split(path, ",")
	for _, part := range parts {
		codeWeight := strings.Split(strings.TrimSpace(part), ":")
		if len(codeWeight) != 2 {
			continue
		}

		code, err := strconv.Atoi(strings.TrimSpace(codeWeight[0]))
		if err != nil {
			continue
		}

		weight, err := strconv.ParseFloat(strings.TrimSpace(codeWeight[1]), 64)
		if err != nil {
			continue
		}

		weights = append(weights, statusWeight{code: code, weight: weight})
	}

	return weights, nil
}

// randFloat64 returns a random float in [0, 1) using crypto/rand (cryptographically secure).
func randFloat64() (float64, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return 0, err
	}
	// Use high 53 bits for uniform [0, 1) like math/rand.Float64
	return float64(binary.BigEndian.Uint64(b)>>11) / (1 << 53), nil
}

// selectStatusCode selects a status code based on weights
func selectStatusCode(weights []statusWeight) int {
	if len(weights) == 0 {
		return http.StatusOK
	}

	if len(weights) == 1 {
		return weights[0].code
	}

	// Calculate total weight
	totalWeight := 0.0
	for _, w := range weights {
		totalWeight += w.weight
	}

	// Generate random number in [0, 1) using crypto/rand and select based on weight
	rf, err := randFloat64()
	if err != nil {
		return weights[0].code
	}
	r := rf * totalWeight
	cumulative := 0.0
	for _, w := range weights {
		cumulative += w.weight
		if r <= cumulative {
			return w.code
		}
	}

	// Fallback to last code
	return weights[len(weights)-1].code
}

// StatusHandler returns a response with the specified status code
func StatusHandler(w http.ResponseWriter, r *http.Request) {
	weights, err := parseStatusCodes(r.URL.Path)
	if err != nil || len(weights) == 0 {
		writeJSONError(w, http.StatusBadRequest, "Invalid status code specification")
		return
	}

	// Select status code (single or weighted random)
	statusCode := selectStatusCode(weights)

	// Validate status code range
	if statusCode < 100 || statusCode > 599 {
		writeJSONError(w, http.StatusBadRequest, "Status code must be between 100 and 599")
		return
	}

	// Send response with the selected status code
	w.WriteHeader(statusCode)
}
