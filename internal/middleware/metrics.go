package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	requestCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "httpbin_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "route", "code"},
	)
	requestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "httpbin_request_duration_seconds",
			Help:    "Request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "route"},
	)
)

// routeLabel normalizes path to a low-cardinality label (e.g. /status/404 -> /status/)
func routeLabel(path string) string {
	path = strings.Trim(path, "/")
	if path == "" {
		return "/"
	}
	parts := strings.SplitN(path, "/", 2)
	if len(parts) == 1 {
		return "/" + parts[0]
	}
	return "/" + parts[0] + "/"
}

// metricsResponseWriter captures status code for metrics
type metricsResponseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (w *metricsResponseWriter) WriteHeader(code int) {
	if !w.written {
		w.statusCode = code
		w.written = true
		w.ResponseWriter.WriteHeader(code)
	}
}

func (w *metricsResponseWriter) Write(b []byte) (int, error) {
	if !w.written {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(b)
}

// Metrics returns middleware that records request count and duration for Prometheus.
func Metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &metricsResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(wrapped, r)
		duration := time.Since(start).Seconds()
		route := routeLabel(r.URL.Path)
		code := strconv.Itoa(wrapped.statusCode)
		requestCount.WithLabelValues(r.Method, route, code).Inc()
		requestDuration.WithLabelValues(r.Method, route).Observe(duration)
	})
}

// MetricsHandler returns the Prometheus /metrics handler (default registry + process metrics).
func MetricsHandler() http.Handler {
	return promhttp.Handler()
}
