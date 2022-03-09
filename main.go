// **** CHANGE package for testing
// service-testattributes_go-otlphttp
// package apm
package main

import (
	"context"
	"log"

	// ***** CHANGE import
	// do not use testing
	// "testing"
	// add os to extract env variable
	"os"

	"go.opentelemetry.io/contrib/propagators/ot"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
)

// ***** CHANGE
// use env variables for OTLP-http endpoint and Service name
var (
	// OTEL_EXPORTER_OTLP_ENDPOINT = os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	OTEL_EXPORTER_OTLP_ENDPOINT = "localhost:4318"
	OTEL_SERVICE_NAME           = os.Getenv("OTEL_SERVICE_NAME")
)

type compositeExporter struct {
	consoleExporter trace.SpanExporter
	atExporter      trace.SpanExporter
}

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

// Initializes tracerProvider with both OTLP-http exporter and console exporter
// Testing purpose
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

	// **** CHANGE
	// use env variable to initialize the tracing, append all options in newclient
	// opts := []otlptracehttp.Option{otlptracehttp.WithInsecure()}
	// 	opts = append(opts, otlptracehttp.WithEndpoint("localhost:4318"))
	// client := otlptracehttp.NewClient(opts...)
	client := otlptracehttp.NewClient(
		otlptracehttp.WithInsecure(),
		otlptracehttp.WithEndpoint(OTEL_EXPORTER_OTLP_ENDPOINT))

	exporter, err := otlptrace.New(context.Background(), client)
	if err != nil {
		panic(err)
	}

	consoleExporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		log.Fatalf("creating stdout exporter: %v", err)
	}

	// **** CHANGE
	//
	// commenting unused struct
	// cExport := compositeExporter{
	// 	consoleExporter: consoleExporter,
	// 	atExporter:      exporter,
	// }
	//
	// V1: create new TracerProvider with a single batchSpanExporter (OTLP-http)
	// tp := trace.NewTracerProvider(
	// 	trace.WithBatcher(exporter),
	// 	trace.WithResource(resources))
	// otel.SetTracerProvider(tp)
	//
	// V2: create new TracerProvider with 2 batchSpanExporters
	tp := trace.NewTracerProvider(
		trace.WithBatcher(consoleExporter),
		trace.WithBatcher(exporter),
		trace.WithResource(resources))
	otel.SetTracerProvider(tp)

	return tp
}

// ****  CHANGE TestAttributes
// do not use testing
// func TestAttributes(t *testing.T) {
func TestAttributes() {
	tp := setUp("active.tax-service.debugattributes.herbert")
	bc := context.Background()

	spanConext, span := otel.Tracer("tax-rate").Start(bc, "pull-tax-rate")
	defer tp.ForceFlush(spanConext)
	span.SetAttributes(attribute.String("key1", "value1"),
		attribute.Bool("boolkey", true),
		attribute.IntSlice("intarraykey", []int{1, 2, 3}))
	defer span.End()
}

// **** CHANGE
// add main for testing
func main() {

	log.Printf("Testing...")

	setUp(OTEL_SERVICE_NAME)

	for i := 0; i < 100; i++ {
		TestAttributes()
	}

	log.Printf("Done.")
}
