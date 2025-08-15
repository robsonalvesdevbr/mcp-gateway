package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

const (
	// ServiceName is the service name for MCP Gateway telemetry
	ServiceName = "docker-mcp-gateway"
	
	// TracerName is the tracer name for MCP Gateway
	TracerName = "github.com/docker/mcp-gateway"
	
	// MeterName is the meter name for MCP Gateway
	MeterName = "github.com/docker/mcp-gateway"
)

var (
	// tracer is the global tracer for MCP Gateway
	tracer trace.Tracer
	
	// meter is the global meter for MCP Gateway
	meter metric.Meter
	
	// ToolCallCounter tracks the number of tool calls with server attribution
	ToolCallCounter metric.Int64Counter
	
	// ToolCallDuration tracks the duration of tool calls in milliseconds
	ToolCallDuration metric.Float64Histogram
	
	// ToolErrorCounter tracks tool call errors by type and server
	ToolErrorCounter metric.Int64Counter
)

// Init initializes the telemetry package with global providers
func Init() {
	// Get tracer from global provider (set by Docker CLI)
	tracer = otel.GetTracerProvider().Tracer(TracerName)
	
	// Get meter from global provider (set by Docker CLI)
	meter = otel.GetMeterProvider().Meter(MeterName)
	
	// Create metrics
	var err error
	
	ToolCallCounter, err = meter.Int64Counter("mcp.tool.calls",
		metric.WithDescription("Number of tool calls executed"),
		metric.WithUnit("1"))
	if err != nil {
		// Log error but don't fail - telemetry should not break the application
		// In production, we'd use proper logging here
	}
	
	ToolCallDuration, err = meter.Float64Histogram("mcp.tool.duration",
		metric.WithDescription("Duration of tool call execution"),
		metric.WithUnit("ms"))
	if err != nil {
		// Log error but don't fail
	}
	
	ToolErrorCounter, err = meter.Int64Counter("mcp.tool.errors",
		metric.WithDescription("Number of tool call errors"),
		metric.WithUnit("1"))
	if err != nil {
		// Log error but don't fail
	}
}

// StartToolCallSpan starts a new span for a tool call with server attribution
func StartToolCallSpan(ctx context.Context, toolName string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	// Add the tool name as a mandatory attribute
	allAttrs := append([]attribute.KeyValue{
		attribute.String("mcp.tool.name", toolName),
	}, attrs...)
	
	return tracer.Start(ctx, "mcp.tool.call",
		trace.WithAttributes(allAttrs...),
		trace.WithSpanKind(trace.SpanKindClient))
}

// StartCommandSpan starts a new span for a command execution
func StartCommandSpan(ctx context.Context, commandPath string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	// Add the command path as an attribute
	allAttrs := append([]attribute.KeyValue{
		attribute.String("mcp.command.path", commandPath),
	}, attrs...)
	
	spanName := "mcp.command." + commandPath
	
	return tracer.Start(ctx, spanName,
		trace.WithAttributes(allAttrs...),
		trace.WithSpanKind(trace.SpanKindServer))
}

// RecordToolError records a tool error with appropriate attributes
func RecordToolError(ctx context.Context, span trace.Span, serverName, serverType, toolName string) {
	if ToolErrorCounter == nil {
		return // Telemetry not initialized
	}
	
	// Record error in span if provided
	if span != nil {
		span.RecordError(nil, trace.WithAttributes(
			attribute.String("mcp.server.name", serverName),
			attribute.String("mcp.server.type", serverType),
		))
	}
	
	ToolErrorCounter.Add(ctx, 1,
		metric.WithAttributes(
			attribute.String("mcp.tool.name", toolName),
			attribute.String("mcp.server.name", serverName),
			attribute.String("mcp.server.type", serverType),
		))
}

// StartPromptSpan starts a new span for a prompt operation
func StartPromptSpan(ctx context.Context, promptName string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	allAttrs := append([]attribute.KeyValue{
		attribute.String("mcp.prompt.name", promptName),
	}, attrs...)
	
	return tracer.Start(ctx, "mcp.prompt.get",
		trace.WithAttributes(allAttrs...),
		trace.WithSpanKind(trace.SpanKindClient))
}

// StartResourceSpan starts a new span for a resource operation
func StartResourceSpan(ctx context.Context, resourceURI string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	allAttrs := append([]attribute.KeyValue{
		attribute.String("mcp.resource.uri", resourceURI),
	}, attrs...)
	
	return tracer.Start(ctx, "mcp.resource.read",
		trace.WithAttributes(allAttrs...),
		trace.WithSpanKind(trace.SpanKindClient))
}

// StartInterceptorSpan starts a new span for interceptor execution
func StartInterceptorSpan(ctx context.Context, when, interceptorType string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	allAttrs := append([]attribute.KeyValue{
		attribute.String("mcp.interceptor.when", when),
		attribute.String("mcp.interceptor.type", interceptorType),
	}, attrs...)
	
	spanName := "mcp.interceptor." + interceptorType
	
	return tracer.Start(ctx, spanName,
		trace.WithAttributes(allAttrs...),
		trace.WithSpanKind(trace.SpanKindInternal))
}