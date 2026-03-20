package caddyroutemetrics

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(RouteMetrics{})
	httpcaddyfile.RegisterHandlerDirective("route_metrics", parseCaddyfile)
	httpcaddyfile.RegisterDirectiveOrder("route_metrics", httpcaddyfile.Before, "respond")
}

type RouteMetrics struct {
	MetricsPath     string          `json:"metrics_path,omitempty"`
	AllowedIPs      []string        `json:"allowed_ips,omitempty"`
	Patterns        []PatternConfig `json:"patterns,omitempty"`
	MaxRoutes       int             `json:"max_routes,omitempty"`
	DurationBuckets []float64       `json:"duration_buckets,omitempty"`
	StatusBuckets   []int           `json:"status_buckets,omitempty"`

	hasBlock    bool
	allowedNets []*net.IPNet
	allowedIPs  []net.IP
	metrics     *metrics
	logger      *zap.Logger
}

type PatternConfig struct {
	Regex       string `json:"regex"`
	Replacement string `json:"replacement"`
}

func (RouteMetrics) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.route_metrics",
		New: func() caddy.Module { return new(RouteMetrics) },
	}
}

func (m *RouteMetrics) Provision(ctx caddy.Context) error {
	m.logger = ctx.Logger()

	if m.MaxRoutes <= 0 {
		m.MaxRoutes = 2000
	}

	// Parse allowed IPs/CIDRs
	if len(m.AllowedIPs) == 0 {
		m.AllowedIPs = []string{"127.0.0.1", "::1"}
	}
	for _, s := range m.AllowedIPs {
		if s == "private_ranges" {
			for _, cidrStr := range privateRanges {
				_, cidr, _ := net.ParseCIDR(cidrStr)
				m.allowedNets = append(m.allowedNets, cidr)
			}
		} else if strings.Contains(s, "/") {
			_, cidr, err := net.ParseCIDR(s)
			if err != nil {
				return err
			}
			m.allowedNets = append(m.allowedNets, cidr)
		} else {
			ip := net.ParseIP(s)
			if ip == nil {
				return fmt.Errorf("invalid IP %q", s)
			}
			m.allowedIPs = append(m.allowedIPs, ip)
		}
	}

	var userPatterns []routePattern
	for _, p := range m.Patterns {
		rp, err := parsePattern(p.Regex, p.Replacement)
		if err != nil {
			return err
		}
		userPatterns = append(userPatterns, rp)
	}

	useDefaults := len(userPatterns) == 0 && !m.hasBlock

	durationBuckets := m.DurationBuckets
	if len(durationBuckets) == 0 {
		durationBuckets = defaultDurationBuckets
	}

	statusBuckets := m.StatusBuckets
	if len(statusBuckets) == 0 {
		statusBuckets = defaultStatusBuckets
	}

	var err error
	m.metrics, err = newMetrics(ctx.GetMetricsRegistry(), metricsConfig{
		patterns:      userPatterns,
		useDefaults:   useDefaults,
		maxRoutes:     m.MaxRoutes,
		buckets:       durationBuckets,
		statusBuckets: statusBuckets,
	})
	if err != nil {
		return err
	}

	m.logger.Info("route_metrics loaded",
		zap.String("metrics_path", m.MetricsPath),
		zap.Int("patterns", len(userPatterns)))
	return nil
}

func (m *RouteMetrics) isAllowed(r *http.Request) bool {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	for _, allowed := range m.allowedIPs {
		if allowed.Equal(ip) {
			return true
		}
	}
	for _, cidr := range m.allowedNets {
		if cidr.Contains(ip) {
			return true
		}
	}
	return false
}

func (m *RouteMetrics) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	reqPath := r.RequestURI
	if idx := strings.IndexByte(reqPath, '?'); idx >= 0 {
		reqPath = reqPath[:idx]
	}

	if m.MetricsPath != "" && reqPath == m.MetricsPath {
		if !m.isAllowed(r) {
			w.WriteHeader(http.StatusForbidden)
			return nil
		}
		m.metrics.handler.ServeHTTP(w, r)
		return nil
	}

	start := time.Now()
	rec := newResponseRecorder(w)
	err := next.ServeHTTP(rec, r)
	durationMs := float64(time.Since(start).Microseconds()) / 1000.0
	m.metrics.recordWithDuration(reqPath, r.Method, rec.status, durationMs)
	return err
}

func (m *RouteMetrics) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	d.Next()

	// Support shorthand: route_metrics /metrics
	if d.NextArg() {
		m.MetricsPath = d.Val()
	}

	for d.NextBlock(0) {
		m.hasBlock = true
		switch d.Val() {
		case "pattern":
			args := d.RemainingArgs()
			if len(args) != 2 {
				return d.ArgErr()
			}
			m.Patterns = append(m.Patterns, PatternConfig{
				Regex:       args[0],
				Replacement: args[1],
			})
		case "duration_buckets":
			args := d.RemainingArgs()
			if len(args) == 0 {
				return d.ArgErr()
			}
			for _, a := range args {
				v, err := strconv.ParseFloat(a, 64)
				if err != nil {
					return d.Errf("invalid duration bucket %q: %v", a, err)
				}
				m.DurationBuckets = append(m.DurationBuckets, v)
			}
		case "status_buckets":
			for _, a := range d.RemainingArgs() {
				v, err := strconv.Atoi(a)
				if err != nil {
					return d.Errf("invalid status bucket %q: %v", a, err)
				}
				m.StatusBuckets = append(m.StatusBuckets, v)
			}
		case "metrics_path":
			if !d.NextArg() {
				return d.ArgErr()
			}
			m.MetricsPath = d.Val()
		case "allowed_ips":
			m.AllowedIPs = d.RemainingArgs()
			if len(m.AllowedIPs) == 0 {
				return d.ArgErr()
			}
		default:
			return d.Errf("unknown subdirective %q", d.Val())
		}
	}

	return nil
}

func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var m RouteMetrics
	err := m.UnmarshalCaddyfile(h.Dispenser)
	return &m, err
}

var privateRanges = []string{
	"127.0.0.0/8",
	"10.0.0.0/8",
	"172.16.0.0/12",
	"192.168.0.0/16",
	"::1/128",
	"fc00::/7",
	"fe80::/10",
}

var (
	_ caddy.Provisioner           = (*RouteMetrics)(nil)
	_ caddyhttp.MiddlewareHandler = (*RouteMetrics)(nil)
	_ caddyfile.Unmarshaler       = (*RouteMetrics)(nil)
)
