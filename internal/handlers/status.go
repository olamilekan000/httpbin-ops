package handlers

import (
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func init() {
	// Initialize random seed
	rand.Seed(time.Now().UnixNano())
}

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

	// Generate random number and select based on weight
	r := rand.Float64() * totalWeight
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
