package apm

import (
	"context"
	"log"
	"testing"

	"go.opentelemetry.io/contrib/propagators/ot"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
)

type compositeExporter struct {
	consoleExporter trace.SpanExporter
	atExporter      trace.SpanExporter
}

const (
	COLLECTOR_URL = "aternity-perf-awplatform.aw.k8sw.dev.activenetwork.com"
)

func (e *compositeExporter) ExportSpans(ctx context.Context, spans []trace.ReadOnlySpan) error {
	if err := e.consoleExporter.ExportSpans(ctx, spans); err != nil {
		return err
	}

	return e.atExporter.ExportSpans(ctx, spans)
}

func (e *compositeExporter) Shutdown(ctx context.Context) error {
	if err := e.consoleExporter.Shutdown(ctx); err != nil {
		return err
	}

	return e.atExporter.Shutdown(ctx)
}

func setUp(serviceName string) *trace.TracerProvider {
	resources, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			attribute.String("service.name", serviceName),
			attribute.String("library.language", "go"),
		),
	)
	if err != nil {
		panic(err)
	}

	otPropagator := ot.OT{}
	otel.SetTextMapPropagator(otPropagator)

	opts := []otlptracehttp.Option{otlptracehttp.WithInsecure()}
	opts = append(opts, otlptracehttp.WithEndpoint(COLLECTOR_URL))
	client := otlptracehttp.NewClient(opts...)
	exporter, err := otlptrace.New(context.Background(), client)
	if err != nil {
		panic(err)
	}

	consoleExporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		log.Fatalf("creating stdout exporter: %v", err)
	}

	cExport := compositeExporter{
		consoleExporter: consoleExporter,
		atExporter:      exporter,
	}

	// Set the main batched tracer provider
	tp := trace.NewTracerProvider(
		trace.WithBatcher(&cExport),
		trace.WithResource(resources),
	)
	otel.SetTracerProvider(tp)

	return tp
}

func TestAttributes(t *testing.T) {
	tp := setUp("active.tax-service.debugattributes")
	bc := context.Background()

	spanConext, span := otel.Tracer("tax-rate").Start(bc, "pull-tax-rate")
	defer tp.ForceFlush(spanConext)
	span.SetAttributes(attribute.String("key1", "value1"),
		attribute.Bool("boolkey", true),
		attribute.IntSlice("intarraykey", []int{1, 2, 3}))
	defer span.End()
}
