package observability

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

// InitTracing configures OpenTelemetry OTLP HTTP tracing (e.g., to Jaeger) and W3C propagation.
// If OTEL_EXPORTER_OTLP_ENDPOINT is unset, it configures propagation with a no-op shutdown.
func InitTracing(ctx context.Context, serviceName string) (func(context.Context) error, error) {
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	endpoint := strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"))
	if endpoint == "" {
		return func(context.Context) error { return nil }, nil
	}

	opts := []otlptracehttp.Option{}
	if strings.HasPrefix(endpoint, "http://") || strings.HasPrefix(endpoint, "https://") {
		opts = append(opts, otlptracehttp.WithEndpointURL(endpoint))
	} else {
		opts = append(opts, otlptracehttp.WithEndpoint(endpoint), otlptracehttp.WithInsecure())
	}

	exporter, err := otlptracehttp.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create OTLP trace exporter: %w", err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(semconv.ServiceName(serviceName)),
	)
	if err != nil {
		return nil, fmt.Errorf("create OTel resource: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)

	return tp.Shutdown, nil
}

// TracingMiddleware instruments incoming Fiber requests with OpenTelemetry server spans,
// extracts W3C traceparent/tracestate headers, stores the span context on c.UserContext(),
// and exposes X-Trace-ID on the response.
func TracingMiddleware(serviceName string) fiber.Handler {
	tracer := otel.Tracer(serviceName)
	return func(c *fiber.Ctx) error {
		carrier := propagation.HeaderCarrier{}
		for k, values := range c.GetReqHeaders() {
			for _, v := range values {
				carrier.Set(k, v)
			}
		}

		parentCtx := otel.GetTextMapPropagator().Extract(c.UserContext(), carrier)
		ctx, span := tracer.Start(
			parentCtx,
			c.Method()+" "+c.Path(),
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				attribute.String("service.name", serviceName),
				attribute.String("http.method", c.Method()),
				attribute.String("url.path", c.Path()),
			),
		)
		defer span.End()

		c.SetUserContext(ctx)
		if sc := span.SpanContext(); sc.IsValid() {
			c.Set("X-Trace-ID", sc.TraceID().String())
		}

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
			span.RecordError(err)
		}
		if statusCode >= 500 {
			span.SetStatus(codes.Error, http.StatusText(statusCode))
		}
		span.SetAttributes(
			attribute.String("http.route", routePath),
			attribute.Int("http.status_code", statusCode),
		)

		return err
	}
}

// TracedRoundTripper wraps an http.RoundTripper to create client spans and propagate W3C trace headers.
type TracedRoundTripper struct {
	Base http.RoundTripper
}

// RoundTrip executes an HTTP request inside an OpenTelemetry client span and injects W3C headers.
func (t *TracedRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	base := t.Base
	if base == nil {
		base = http.DefaultTransport
	}

	ctx := req.Context()
	tracer := otel.Tracer("http-client")
	ctx, span := tracer.Start(
		ctx,
		req.Method+" "+req.URL.Host+req.URL.Path,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("http.method", req.Method),
			attribute.String("url.full", req.URL.String()),
		),
	)
	defer span.End()

	cloned := req.Clone(ctx)
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(cloned.Header))

	resp, err := base.RoundTrip(cloned)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	span.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))
	if resp.StatusCode >= 500 {
		span.SetStatus(codes.Error, http.StatusText(resp.StatusCode))
	}
	return resp, nil
}

// NewTracedHTTPClient returns an *http.Client that records OpenTelemetry client spans
// and propagates W3C traceparent headers.
func NewTracedHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout:   timeout,
		Transport: &TracedRoundTripper{Base: http.DefaultTransport},
	}
}
