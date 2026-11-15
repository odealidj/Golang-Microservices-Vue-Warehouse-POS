package observability

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestRegisterMetrics_ExposesPrometheusMetrics(t *testing.T) {
	app := fiber.New()
	RegisterMetrics(app, "metrics-test-service")
	app.Get("/ping", func(c *fiber.Ctx) error {
		return c.SendString("pong")
	})

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/ping", nil))
	if err != nil {
		t.Fatalf("/ping failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 on /ping, got %d", resp.StatusCode)
	}

	metricsResp, err := app.Test(httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if err != nil {
		t.Fatalf("/metrics failed: %v", err)
	}
	body, _ := io.ReadAll(metricsResp.Body)
	text := string(body)
	if !strings.Contains(text, `http_requests_total{method="GET",path="/ping",service="metrics-test-service",status="200"}`) {
		t.Fatalf("expected http_requests_total metric in /metrics output, got:\n%s", text)
	}
	if !strings.Contains(text, `http_request_duration_seconds_count{method="GET",path="/ping",service="metrics-test-service"}`) {
		t.Fatalf("expected http_request_duration_seconds metric in /metrics output")
	}
}
