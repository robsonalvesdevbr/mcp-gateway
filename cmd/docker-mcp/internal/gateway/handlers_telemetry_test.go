package gateway

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	mcpclient "github.com/mark3labs/mcp-go/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/docker/mcp-gateway/cmd/docker-mcp/internal/catalog"
	"github.com/docker/mcp-gateway/cmd/docker-mcp/internal/telemetry"
)

// mockMCPClient implements a minimal MCP client for testing
type mockMCPClient struct {
	callToolFunc func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)
	calls        []mcp.CallToolRequest
}

func (m *mockMCPClient) CallTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	m.calls = append(m.calls, request)
	if m.callToolFunc != nil {
		return m.callToolFunc(ctx, request)
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{{
			Type: "text",
			Text: "mock result",
		}},
	}, nil
}

func (m *mockMCPClient) GetPrompt(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	return nil, nil
}

func (m *mockMCPClient) ReadResource(ctx context.Context, request mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	return nil, nil
}

func (m *mockMCPClient) Initialize(ctx context.Context, request mcp.InitializeRequest, debug bool) (*mcp.InitializeResult, error) {
	return nil, nil
}

func (m *mockMCPClient) ListTools(ctx context.Context) ([]mcp.Tool, error) {
	return nil, nil
}

func (m *mockMCPClient) ListPrompts(ctx context.Context) ([]mcp.Prompt, error) {
	return nil, nil
}

func (m *mockMCPClient) ListResources(ctx context.Context) ([]mcp.Resource, error) {
	return nil, nil
}

func (m *mockMCPClient) ListResourceTemplates(ctx context.Context) ([]mcp.ResourceTemplate, error) {
	return nil, nil
}

func (m *mockMCPClient) Close() error {
	return nil
}

// mockClientPool for testing
type mockClientPool struct {
	client       mcpclient.Client
	acquireError error
	acquireCalls int
	releaseCalls int
}

func (m *mockClientPool) AcquireClient(ctx context.Context, serverConfig ServerConfig, readOnly *bool) (mcpclient.Client, error) {
	m.acquireCalls++
	if m.acquireError != nil {
		return nil, m.acquireError
	}
	return m.client, nil
}

func (m *mockClientPool) ReleaseClient(client mcpclient.Client) {
	m.releaseCalls++
}

func (m *mockClientPool) runToolContainer(ctx context.Context, tool catalog.Tool, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return nil, nil
}

func (m *mockClientPool) Close() {
}

// setupTestTelemetry creates test providers with in-memory exporters
func setupTestTelemetry(t *testing.T) (*tracetest.SpanRecorder, *sdkmetric.ManualReader) {
	// Create in-memory span recorder for traces
	spanRecorder := tracetest.NewSpanRecorder()
	tracerProvider := trace.NewTracerProvider(
		trace.WithSpanProcessor(spanRecorder),
	)
	otel.SetTracerProvider(tracerProvider)

	// Create manual reader for metrics
	reader := sdkmetric.NewManualReader()
	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(reader),
	)
	otel.SetMeterProvider(meterProvider)

	// Initialize telemetry package
	telemetry.Init()

	// Cleanup
	t.Cleanup(func() {
		// Reset to noop providers after test
		otel.SetTracerProvider(trace.NewTracerProvider())
		otel.SetMeterProvider(sdkmetric.NewMeterProvider())
	})

	return spanRecorder, reader
}

func TestMcpServerToolHandler_Telemetry(t *testing.T) {
	spanRecorder, metricReader := setupTestTelemetry(t)

	// Create mock client
	mockClient := &mockMCPClient{
		callToolFunc: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// Simulate some processing time
			time.Sleep(50 * time.Millisecond)
			return &mcp.CallToolResult{
				Content: []mcp.Content{{
					Type: "text",
					Text: "success",
				}},
			}, nil
		},
	}

	// Create mock pool
	mockPool := &mockClientPool{
		client: mockClient,
	}

	// Create gateway with mock pool
	g := &Gateway{
		clientPool: mockPool,
	}

	// Create server config
	serverConfig := ServerConfig{
		Name: "test-server",
		Spec: catalog.Server{
			Image: "test/server:latest",
		},
	}

	// Create the handler
	handler := g.mcpServerToolHandler(serverConfig, mcp.ToolAnnotation{})

	// Create request
	request := mcp.CallToolRequest{
		Name: "test_tool",
		Arguments: map[string]interface{}{
			"arg1": "value1",
		},
	}

	// Execute the handler
	ctx := context.Background()
	result, err := handler(ctx, request)

	// Verify no error and result is correct
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Content, 1)
	assert.Equal(t, "success", result.Content[0].Text)

	// Verify spans were created
	spans := spanRecorder.Ended()
	require.Len(t, spans, 1, "should have created one span")
	
	span := spans[0]
	assert.Equal(t, "mcp.tool.call", span.Name())
	
	// Check span attributes
	attrs := span.Attributes()
	attrMap := make(map[string]string)
	for _, attr := range attrs {
		attrMap[string(attr.Key)] = attr.Value.AsString()
	}
	
	assert.Equal(t, "test_tool", attrMap["mcp.tool.name"])
	assert.Equal(t, "test-server", attrMap["mcp.server.origin"])
	assert.Equal(t, "docker", attrMap["mcp.server.type"]) // Inferred from Image field

	// Verify metrics were recorded
	var rm metricdata.ResourceMetrics
	err = metricReader.Collect(ctx, &rm)
	require.NoError(t, err)

	// Check tool call counter
	foundCounter := false
	foundDuration := false
	
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			switch m.Name {
			case "mcp.tool.calls":
				foundCounter = true
				sum := m.Data.(metricdata.Sum[int64])
				assert.Equal(t, int64(1), sum.DataPoints[0].Value)
				
				// Check attributes
				attrs := sum.DataPoints[0].Attributes
				toolName, _ := attrs.Value(attribute.Key("mcp.tool.name"))
				assert.Equal(t, "test_tool", toolName.AsString())
				
				serverOrigin, _ := attrs.Value(attribute.Key("mcp.server.origin"))
				assert.Equal(t, "test-server", serverOrigin.AsString())
				
			case "mcp.tool.duration":
				foundDuration = true
				hist := m.Data.(metricdata.Histogram[float64])
				assert.Greater(t, hist.DataPoints[0].Count, uint64(0))
				// Should be at least 50ms due to our sleep
				assert.GreaterOrEqual(t, hist.DataPoints[0].Sum, float64(50))
			}
		}
	}
	
	assert.True(t, foundCounter, "tool call counter should be recorded")
	assert.True(t, foundDuration, "tool duration should be recorded")

	// Verify client pool was used correctly
	assert.Equal(t, 1, mockPool.acquireCalls, "should acquire client once")
	assert.Equal(t, 1, mockPool.releaseCalls, "should release client once")
}

func TestMcpServerToolHandler_ErrorTelemetry(t *testing.T) {
	spanRecorder, metricReader := setupTestTelemetry(t)

	testError := errors.New("tool execution failed")

	// Create mock client that returns an error
	mockClient := &mockMCPClient{
		callToolFunc: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return nil, testError
		},
	}

	// Create mock pool
	mockPool := &mockClientPool{
		client: mockClient,
	}

	// Create gateway with mock pool
	g := &Gateway{
		clientPool: mockPool,
	}

	// Create server config
	serverConfig := ServerConfig{
		Name: "test-server",
		Spec: catalog.Server{
			SSEEndpoint: "http://example.com/sse", // This makes it an SSE type
		},
	}

	// Create the handler
	handler := g.mcpServerToolHandler(serverConfig, mcp.ToolAnnotation{})

	// Create request
	request := mcp.CallToolRequest{
		Name: "failing_tool",
	}

	// Execute the handler
	ctx := context.Background()
	result, err := handler(ctx, request)

	// Verify error is returned
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, testError, err)

	// Verify span recorded the error
	spans := spanRecorder.Ended()
	require.Len(t, spans, 1)
	
	span := spans[0]
	assert.Equal(t, "mcp.tool.call", span.Name())
	
	// Check that span recorded error
	events := span.Events()
	hasError := false
	for _, event := range events {
		if event.Name == "exception" {
			hasError = true
			break
		}
	}
	assert.True(t, hasError, "span should record error event")

	// Check span attributes include server type
	attrs := span.Attributes()
	attrMap := make(map[string]string)
	for _, attr := range attrs {
		attrMap[string(attr.Key)] = attr.Value.AsString()
	}
	assert.Equal(t, "sse", attrMap["mcp.server.type"])

	// Verify error metrics were recorded
	var rm metricdata.ResourceMetrics
	err = metricReader.Collect(ctx, &rm)
	require.NoError(t, err)

	// Check error counter
	foundErrors := false
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name == "mcp.tool.errors" {
				foundErrors = true
				sum := m.Data.(metricdata.Sum[int64])
				assert.Equal(t, int64(1), sum.DataPoints[0].Value)
				
				// Check attributes
				attrs := sum.DataPoints[0].Attributes
				toolName, _ := attrs.Value(attribute.Key("mcp.tool.name"))
				assert.Equal(t, "failing_tool", toolName.AsString())
				
				serverOrigin, _ := attrs.Value(attribute.Key("mcp.server.origin"))
				assert.Equal(t, "test-server", serverOrigin.AsString())
			}
		}
	}
	
	assert.True(t, foundErrors, "error counter should be recorded")
}

func TestMcpServerToolHandler_AcquireClientError(t *testing.T) {
	spanRecorder, metricReader := setupTestTelemetry(t)

	acquireError := errors.New("failed to acquire client")

	// Create mock pool that fails to acquire
	mockPool := &mockClientPool{
		acquireError: acquireError,
	}

	// Create gateway with mock pool
	g := &Gateway{
		clientPool: mockPool,
	}

	// Create server config
	serverConfig := ServerConfig{
		Name: "test-server",
		Spec: catalog.Server{
			Command: []string{"mcp", "server"}, // This makes it a stdio type
		},
	}

	// Create the handler
	handler := g.mcpServerToolHandler(serverConfig, mcp.ToolAnnotation{})

	// Create request
	request := mcp.CallToolRequest{
		Name: "test_tool",
	}

	// Execute the handler
	ctx := context.Background()
	result, err := handler(ctx, request)

	// Verify error is returned
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, acquireError, err)

	// Verify span recorded the error
	spans := spanRecorder.Ended()
	require.Len(t, spans, 1)
	
	span := spans[0]
	
	// Check span attributes include server type
	attrs := span.Attributes()
	attrMap := make(map[string]string)
	for _, attr := range attrs {
		attrMap[string(attr.Key)] = attr.Value.AsString()
	}
	assert.Equal(t, "stdio", attrMap["mcp.server.type"])

	// Verify error metrics
	var rm metricdata.ResourceMetrics
	err = metricReader.Collect(ctx, &rm)
	require.NoError(t, err)

	foundErrors := false
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name == "mcp.tool.errors" {
				foundErrors = true
				sum := m.Data.(metricdata.Sum[int64])
				assert.Equal(t, int64(1), sum.DataPoints[0].Value)
			}
		}
	}
	
	assert.True(t, foundErrors, "error counter should be recorded")
	
	// Verify pool was called but client was never released (since acquire failed)
	assert.Equal(t, 1, mockPool.acquireCalls)
	assert.Equal(t, 0, mockPool.releaseCalls)
}