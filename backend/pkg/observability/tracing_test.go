package observability

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

func TestInitTracing_NoEndpointReturnsNoopShutdown(t *testing.T) {
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")
	shutdown, err := InitTracing(context.Background(), "test-service")
	if err != nil {
		t.Fatalf("InitTracing failed: %v", err)
	}
	if err := shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown failed: %v", err)
	}
}

func TestTracingMiddleware_PropagatesTraceContext(t *testing.T) {
	tp := sdktrace.NewTracerProvider()
	prev := otel.GetTracerProvider()
	otel.SetTracerProvider(tp)
	defer func() {
		otel.SetTracerProvider(prev)
		_ = tp.Shutdown(context.Background())
	}()

	_, err := InitTracing(context.Background(), "test-service")
	if err != nil {
		t.Fatalf("InitTracing failed: %v", err)
	}

	app := fiber.New()
	app.Use(TracingMiddleware("test-service"))
	app.Get("/ping", func(c *fiber.Ctx) error {
		if !trace.SpanFromContext(c.UserContext()).SpanContext().IsValid() {
			return c.Status(fiber.StatusInternalServerError).SendString("missing span")
		}
		return c.SendString("pong")
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("/ping failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 on /ping, got %d", resp.StatusCode)
	}
	if resp.Header.Get("X-Trace-ID") == "" {
		t.Fatal("expected X-Trace-ID header on response")
	}
}

func TestTracedHTTPClient_InjectsTraceparentHeader(t *testing.T) {
	tp := sdktrace.NewTracerProvider()
	prev := otel.GetTracerProvider()
	otel.SetTracerProvider(tp)
	defer func() {
		otel.SetTracerProvider(prev)
		_ = tp.Shutdown(context.Background())
	}()
	_, _ = InitTracing(context.Background(), "client-test")

	var gotTraceparent string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotTraceparent = r.Header.Get("Traceparent")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := NewTracedHTTPClient(5 * time.Second)
	ctx, span := otel.Tracer("test").Start(context.Background(), "parent")
	defer span.End()

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+"/test", nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("client.Do failed: %v", err)
	}
	_ = resp.Body.Close()

	if gotTraceparent == "" {
		t.Fatal("expected Traceparent header to be propagated to downstream server")
	}
}
