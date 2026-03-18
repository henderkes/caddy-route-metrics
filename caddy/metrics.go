package caddyroutemetrics

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type metrics struct {
	duration *prometheus.HistogramVec
	total    *prometheus.CounterVec
	errors   *prometheus.CounterVec
	handler  http.Handler

	patterns      []routePattern
	useDefaults   bool
	statusBuckets map[int]string
	maxRoutes     int
	routesSeen    sync.Map
	routesCount   atomic.Int64
}

type metricsConfig struct {
	patterns      []routePattern
	useDefaults   bool
	maxRoutes     int
	buckets       []float64
	statusBuckets []int
}

type routePattern struct {
	re          *regexp.Regexp
	replacement string
}

var defaultDurationBuckets = []float64{5, 10, 25, 35, 50, 75, 100, 150, 250, 500, 1000}

var defaultStatusBuckets []int

var defaultSegmentPatterns = []struct {
	re          *regexp.Regexp
	replacement string
}{
	{regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`), ":uuid"},
	{regexp.MustCompile(`^[0-9a-f]{24,64}$`), ":id"},
	{regexp.MustCompile(`^\d+$`), ":id"},
}

func parsePattern(pattern, replacement string) (routePattern, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return routePattern{}, fmt.Errorf("invalid pattern %q: %w", pattern, err)
	}
	return routePattern{re: re, replacement: replacement}, nil
}

func buildStatusBuckets(codes []int) map[int]string {
	m := make(map[int]string, len(codes))
	for _, c := range codes {
		m[c] = strconv.Itoa(c)
	}
	return m
}

func newMetrics(registry prometheus.Registerer, cfg metricsConfig) (*metrics, error) {
	// Build a handler that can serve the metrics from this registry
	var gatherer prometheus.Gatherer
	if g, ok := registry.(prometheus.Gatherer); ok {
		gatherer = g
	} else {
		gatherer = prometheus.DefaultGatherer
	}

	m := &metrics{
		patterns:      cfg.patterns,
		useDefaults:   cfg.useDefaults,
		statusBuckets: buildStatusBuckets(cfg.statusBuckets),
		maxRoutes:     cfg.maxRoutes,
	}

	m.duration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "caddy",
		Subsystem: "route",
		Name:      "request_duration_milliseconds",
		Help:      "Request duration in milliseconds per route.",
		Buckets:   cfg.buckets,
	}, []string{"route", "method", "status_code"})

	m.total = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "caddy",
		Subsystem: "route",
		Name:      "requests_total",
		Help:      "Total requests per route.",
	}, []string{"route", "method", "status_code"})

	m.errors = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "caddy",
		Subsystem: "route",
		Name:      "errors_total",
		Help:      "Total 5xx responses per route.",
	}, []string{"route", "method", "status_code"})

	for _, c := range []prometheus.Collector{m.duration, m.total, m.errors} {
		if err := registry.Register(c); err != nil {
			if _, ok := err.(prometheus.AlreadyRegisteredError); !ok {
				return nil, fmt.Errorf("route_metrics: %w", err)
			}
		}
	}

	m.handler = promhttp.HandlerFor(gatherer, promhttp.HandlerOpts{})

	return m, nil
}

func (m *metrics) resolveRoute(uri string) string {
	if idx := strings.IndexByte(uri, '?'); idx >= 0 {
		uri = uri[:idx]
	}

	matched := false
	for _, p := range m.patterns {
		if p.re.MatchString(uri) {
			uri = p.re.ReplaceAllString(uri, p.replacement)
			matched = true
		}
	}

	if !m.useDefaults {
		if !matched {
			return ""
		}
		return uri
	}

	parts := strings.Split(uri, "/")
	changed := false
	for i, part := range parts {
		if part == "" {
			continue
		}
		for _, dp := range defaultSegmentPatterns {
			if dp.re.MatchString(part) {
				parts[i] = dp.replacement
				changed = true
				break
			}
		}
	}
	if changed {
		return strings.Join(parts, "/")
	}

	return uri
}

func (m *metrics) formatStatus(code int) string {
	if s, ok := m.statusBuckets[code]; ok {
		return s
	}
	switch {
	case code < 200:
		return "1xx"
	case code < 300:
		return "2xx"
	case code < 400:
		return "3xx"
	case code < 500:
		return "4xx"
	default:
		return "5xx"
	}
}

func (m *metrics) recordWithDuration(path, method string, statusCode int, durationMs float64) {
	route := m.resolveRoute(path)
	if route == "" {
		return
	}

	if _, loaded := m.routesSeen.LoadOrStore(route, struct{}{}); !loaded {
		if m.routesCount.Add(1) > int64(m.maxRoutes) {
			m.routesSeen.Delete(route)
			m.routesCount.Add(-1)
			route = "__overflow__"
		}
	}

	sc := m.formatStatus(statusCode)

	m.duration.WithLabelValues(route, method, sc).Observe(durationMs)
	m.total.WithLabelValues(route, method, sc).Inc()

	if statusCode >= 500 {
		m.errors.WithLabelValues(route, method, sc).Inc()
	}
}
