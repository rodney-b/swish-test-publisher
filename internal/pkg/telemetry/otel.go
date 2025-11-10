package telemetry

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"

	"github.com/rodney-b/swish-test-publisher/internal/pkg/config"
	"github.com/rodney-b/swish-test-publisher/pkg/certs"
)

// initOTel bootstraps the OpenTelemetry pipeline.
// If it does not return an error, make sure to call shutdown for proper cleanup.
// Taken from https://opentelemetry.io/docs/languages/go/getting-started/
// Seems solid enough so it was used with barely any changes
func initOTel(ctx context.Context, cp config.ConfigProvider) (func(context.Context) error, error) {
	var shutdownFuncs []func(context.Context) error
	var err error

	// shutdown calls cleanup functions registered via shutdownFuncs.
	// The errors from the calls are joined.
	// Each registered cleanup will be invoked once.
	shutdown := func(ctx context.Context) error {
		var shutdownErr error
		for _, fn := range shutdownFuncs {
			shutdownErr = errors.Join(shutdownErr, fn(ctx))
		}
		shutdownFuncs = nil
		return shutdownErr
	}

	// handleErr calls shutdown for cleanup and makes sure that all errors are returned.
	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}

	// Set up propagator.
	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	// Set up trace provider.
	tracerProvider, err := newTracerProvider(ctx, cp)
	if err != nil {
		handleErr(err)
		return shutdown, err
	}
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)

	// Set up meter provider.
	meterProvider, err := newMeterProvider(ctx, cp)
	if err != nil {
		handleErr(err)
		return shutdown, err
	}
	shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
	otel.SetMeterProvider(meterProvider)

	return shutdown, nil
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func newMeterProvider(ctx context.Context, cp config.ConfigProvider) (*metric.MeterProvider, error) {
	var exporter metric.Exporter
	var err error

	if cp.GetOtelStdoutExporterEnabled() {
		exporter, err = stdoutmetric.New()
	} else {
		// TODO: Replace data source issuer with a dedicated issuer
		tlsConfig, err := certs.CreateTLSConfig(cp.GetDataSourceClientCA(), cp.GetDataSourceClientCert(), cp.GetDataSourceClientCertKey())
		if err != nil {
			return nil, err
		}

		exporter, err = otlpmetrichttp.New(
			ctx,
			otlpmetrichttp.WithEndpoint(cp.GetOTelHTTPReceiverURL()),
			otlpmetrichttp.WithTLSClientConfig(tlsConfig),
		)
	}
	if err != nil {
		return nil, err
	}

	// Default interval between exports is 1m but it can be changed with the WithInterval option
	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(exporter)),
	)
	return meterProvider, nil
}

func newTracerProvider(ctx context.Context, cp config.ConfigProvider) (*trace.TracerProvider, error) {
	traceExporter, err := otlptracehttp.New(
		ctx,
		otlptracehttp.WithEndpoint(cp.GetOTelHTTPReceiverURL()),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	// Default interval between exports is 5s but it can be changed with the WithInterval option
	// as well as timeout options
	tracerProvider := trace.NewTracerProvider(
		trace.WithBatcher(traceExporter),
	)
	return tracerProvider, nil
}
