package main

// OpenTelemetry: the API sends traces, metrics and logs to a collector over OTLP.
//
// Everything is configured with the standard OTEL_* environment variables:
//
//	OTEL_EXPORTER_OTLP_ENDPOINT  collector URL, e.g. http://lgtm:4318 (unset: telemetry off)
//	OTEL_SERVICE_NAME            name shown in Grafana, e.g. seminar-api
//	OTEL_METRIC_EXPORT_INTERVAL  milliseconds between metric exports (default 60000)

import (
	"context"
	"log/slog"
	"os"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// setupTelemetry starts the three exporters. Without an endpoint it does
// nothing, so the API runs exactly as before.
//
// Data is sent in batches every few seconds. This API has no graceful shutdown
// (step 3 adds it), so the last few seconds are lost when the container stops:
// fine for a demo.
func setupTelemetry(ctx context.Context) error {
	if os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") == "" {
		return nil
	}

	// The exporters read the endpoint from the environment.
	traceExporter, err := otlptracehttp.New(ctx)
	if err != nil {
		return err
	}
	metricExporter, err := otlpmetrichttp.New(ctx)
	if err != nil {
		return err
	}
	logExporter, err := otlploghttp.New(ctx)
	if err != nil {
		return err
	}

	// Traces: one span per request, sent in batches.
	tracerProvider := sdktrace.NewTracerProvider(sdktrace.WithBatcher(traceExporter))
	otel.SetTracerProvider(tracerProvider)
	// Pass trace IDs along in HTTP headers (W3C traceparent), so calls across services join one trace.
	otel.SetTextMapPropagator(propagation.TraceContext{})

	// Metrics: counters and histograms, exported every OTEL_METRIC_EXPORT_INTERVAL.
	meterProvider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter)))
	otel.SetMeterProvider(meterProvider)

	// Logs: every slog line still goes to stdout and now also to the collector,
	// tagged with the trace ID of the request that wrote it.
	loggerProvider := sdklog.NewLoggerProvider(sdklog.WithProcessor(sdklog.NewBatchProcessor(logExporter)))
	otel.SetLoggerProvider(loggerProvider)
	slog.SetDefault(slog.New(slog.NewMultiHandler(
		slog.Default().Handler(),
		otelslog.NewHandler("seminar-api"),
	)))

	return nil
}
