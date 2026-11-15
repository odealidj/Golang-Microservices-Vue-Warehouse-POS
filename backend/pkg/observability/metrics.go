package observability

import (
	"strconv"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	metricsOnce         sync.Once
	httpRequestsTotal   *prometheus.CounterVec
	httpRequestDuration *prometheus.HistogramVec
)

func initMetrics() {
	metricsOnce.Do(func() {
		httpRequestsTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests processed by the service.",
			},
			[]string{"service", "method", "path", "status"},
		)
		httpRequestDuration = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request latency distribution in seconds.",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"service", "method", "path"},
		)
		prometheus.MustRegister(httpRequestsTotal, httpRequestDuration)
	})
}

// RegisterMetrics mounts GET /metrics for Prometheus scraping and records HTTP request
// count and latency histograms per route.
func RegisterMetrics(app *fiber.App, serviceName string) {
	initMetrics()

	app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))

	app.Use(func(c *fiber.Ctx) error {
		if c.Path() == "/metrics" {
			return c.Next()
		}

		start := time.Now()
		err := c.Next()

		routePath := c.Path()
		if r := c.Route(); r != nil && r.Path != "" {
			routePath = r.Path
		}
		statusCode := c.Response().StatusCode()
		if err != nil {
			if fe, ok := err.(*fiber.Error); ok {
				statusCode = fe.Code
			} else if statusCode < 400 {
				statusCode = fiber.StatusInternalServerError
			}
		}

		httpRequestsTotal.WithLabelValues(serviceName, c.Method(), routePath, strconv.Itoa(statusCode)).Inc()
		httpRequestDuration.WithLabelValues(serviceName, c.Method(), routePath).Observe(time.Since(start).Seconds())

		return err
	})
}
