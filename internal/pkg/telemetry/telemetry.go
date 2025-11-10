package telemetry

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/rodney-b/swish-test-publisher/internal/pkg/config"
)

type Telemetry struct {
	logger   *slog.Logger
	meter    metric.Meter
	tracer   trace.Tracer
	Shutdown func()
}

var (
	messageCounter metric.Int64Counter
)

func NewTelemetry(ctx context.Context, cp config.ConfigProvider, logger *slog.Logger) (*Telemetry, error) {
	shutdownFunc, err := initOTel(ctx, cp)
	if err != nil {
		return nil, err
	}

	shutdown := func() {
		err = shutdownFunc(ctx)
		if err != nil {
			logger.Error("error shutting down the telemetry pipeline", "error", err.Error())
		}
	}

	meter := otel.Meter(cp.GetAppName())
	tracer := otel.Tracer(cp.GetAppName())

	err = initMeterInstruments(meter)
	if err != nil {
		return nil, err
	}

	// TODO: init tracer instruments

	tel := Telemetry{
		logger:   logger,
		meter:    meter,
		tracer:   tracer,
		Shutdown: shutdown,
	}

	return &tel, nil
}

func initMeterInstruments(meter metric.Meter) error {
	var err error

	messageCounter, err = meter.Int64Counter(
		"published.message",
		metric.WithDescription("count of messages published"),
	)
	if err != nil {
		return err
	}

	return nil
}

func (tel *Telemetry) IncrementMessageCounter(ctx context.Context, cp config.ConfigProvider) {
	mode := "one"
	if cp.GetAppMode() == 2 {
		mode = "two"
	}

	metricAttrs := []attribute.KeyValue{
		attribute.String("mode", mode),
	}

	messageCounter.Add(ctx, 1, metric.WithAttributes(metricAttrs...))
}
