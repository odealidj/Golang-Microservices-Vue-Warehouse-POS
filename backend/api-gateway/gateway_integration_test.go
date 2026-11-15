//go:build integration

package main

import (
	"context"
	"encoding/json"
	"io"
	"micro-warehouse/api-gateway/middleware"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func newTestGatewayConfig(upstreamURL string) Config {
	return Config{
		Port: "0",
		Services: map[string]ServiceConfig{
			"user":         {Name: "user-service", URL: upstreamURL},
			"role":         {Name: "role-service", URL: upstreamURL},
			"assign-role":  {Name: "assign-role-service", URL: upstreamURL},
			"auth":         {Name: "auth-service", URL: upstreamURL},
			"product":      {Name: "product-service", URL: upstreamURL},
			"merchant":     {Name: "merchant-service", URL: upstreamURL},
			"notification": {Name: "notification-service", URL: upstreamURL},
			"transaction":  {Name: "transaction-service", URL: upstreamURL},
			"midtrans":     {Name: "midtrans-service", URL: upstreamURL},
			"warehouse":    {Name: "warehouse-service", URL: upstreamURL},
		},
	}
}

func TestGatewayIntegration_HealthJWTHeaderSanitizationAndMidtransWebhook(t *testing.T) {
	tp := sdktrace.NewTracerProvider()
	prevTP := otel.GetTracerProvider()
	prevProp := otel.GetTextMapPropagator()
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	defer func() {
		otel.SetTracerProvider(prevTP)
		otel.SetTextMapPropagator(prevProp)
		_ = tp.Shutdown(context.Background())
	}()

	var capturedHeaders http.Header
	var capturedPath string

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedHeaders = r.Header.Clone()
		capturedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer upstream.Close()

	jwtCfg := middleware.JWTConfig{
		SecretKey: "integration-test-secret",
		Issuer:    "warehouse-api-gateway",
		Duration:  time.Hour,
	}

	app := buildApp(newTestGatewayConfig(upstream.URL), jwtCfg, nil)

	// 1. Health endpoint
	healthResp, err := app.Test(httptest.NewRequest(http.MethodGet, "/health", nil))
	if err != nil {
		t.Fatalf("health request failed: %v", err)
	}
	if healthResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on /health, got %d", healthResp.StatusCode)
	}

	// 2. Protected route without JWT -> 401
	unauthResp, err := app.Test(httptest.NewRequest(http.MethodGet, "/api/v1/products", nil))
	if err != nil {
		t.Fatalf("unauth request failed: %v", err)
	}
	if unauthResp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 without token, got %d", unauthResp.StatusCode)
	}

	// 3. Protected route with valid JWT + spoofed identity headers -> headers sanitized & replaced from JWT
	token, err := middleware.GenerateJWT(42, "manager@mail.com", "Manager", jwtCfg)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/products?page=1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-User-ID", "999999")
	req.Header.Set("X-User-Email", "attacker@evil.com")
	req.Header.Set("X-User-Roles", "SuperAdmin")

	authResp, err := app.Test(req)
	if err != nil {
		t.Fatalf("authenticated request failed: %v", err)
	}
	if authResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(authResp.Body)
		t.Fatalf("expected 200 for valid JWT, got %d: %s", authResp.StatusCode, string(body))
	}
	if capturedPath != "/api/v1/products" {
		t.Fatalf("expected upstream path /api/v1/products, got %s", capturedPath)
	}
	if got := capturedHeaders.Get("X-User-ID"); got != "42" {
		t.Fatalf("expected X-User-ID=42 from JWT, got %q", got)
	}
	if got := capturedHeaders.Get("X-User-Email"); got != "manager@mail.com" {
		t.Fatalf("expected X-User-Email=manager@mail.com, got %q", got)
	}
	if got := capturedHeaders.Get("X-Gateway"); got != "warehouse-api-gateway" {
		t.Fatalf("expected X-Gateway header, got %q", got)
	}
	if authResp.Header.Get("X-Trace-ID") == "" {
		t.Fatal("expected X-Trace-ID header on response")
	}
	if capturedHeaders.Get("Traceparent") == "" {
		t.Fatal("expected Traceparent header to be propagated to upstream service")
	}

	// 4. Public /metrics endpoint -> 200 with Prometheus metrics
	metricsResp, err := app.Test(httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if err != nil {
		t.Fatalf("metrics request failed: %v", err)
	}
	if metricsResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on /metrics, got %d", metricsResp.StatusCode)
	}
	metricsBody, _ := io.ReadAll(metricsResp.Body)
	if !strings.Contains(string(metricsBody), "http_requests_total") {
		t.Fatalf("expected http_requests_total in /metrics output")
	}

	// 5. Public Midtrans webhook route -> forwards without JWT
	webhookReq := httptest.NewRequest(http.MethodPost, "/api/v1/midtrans/callback", strings.NewReader(`{"order_id":"ORDER-1"}`))
	webhookReq.Header.Set("Content-Type", "application/json")
	webhookResp, err := app.Test(webhookReq)
	if err != nil {
		t.Fatalf("webhook request failed: %v", err)
	}
	if webhookResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on midtrans webhook, got %d", webhookResp.StatusCode)
	}
	var parsed map[string]bool
	if err := json.NewDecoder(webhookResp.Body).Decode(&parsed); err != nil || !parsed["ok"] {
		t.Fatalf("unexpected webhook response body: err=%v parsed=%v", err, parsed)
	}
}
