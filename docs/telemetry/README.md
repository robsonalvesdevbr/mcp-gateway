# MCP Gateway Telemetry Documentation

## Overview

The MCP Gateway integrates with Docker CLI's OpenTelemetry (OTEL) infrastructure to provide comprehensive observability for MCP operations. This telemetry helps track tool calls, measure performance, and debug issues across distributed MCP server interactions.

## Architecture

The MCP Gateway telemetry follows a **provider inheritance model**:

```
Docker CLI
├─> Reads OTEL configuration from Docker context or environment
├─> Creates global TracerProvider and MeterProvider with configured exporters
├─> Sets providers via otel.SetTracerProvider/SetMeterProvider
└─> MCP Gateway Plugin
    ├─> Inherits providers via otel.GetTracerProvider/GetMeterProvider
    ├─> Creates traces and metrics using inherited providers
    └─> Telemetry flows to Docker CLI's configured endpoint
```

### Key Design Principles

1. **No Custom Exporters**: MCP Gateway never creates its own exporters or endpoints
2. **Provider Inheritance**: Uses Docker CLI's global OTEL providers
3. **Transparent Configuration**: Respects Docker context and environment settings
4. **Event-Driven**: Telemetry is only generated when operations occur

## Instrumentation Scope

### Traces

- **Scope**: `github.com/docker/mcp-gateway`
- **Spans Created**:
  - `mcp.tool.call` - Tool execution spans with server attribution
  - `mcp.command.*` - Command execution spans (future)
  - `mcp.prompt.get` - Prompt retrieval spans (future)
  - `mcp.resource.read` - Resource reading spans (future)

### Metrics

- **Scope**: `github.com/docker/mcp-gateway`
- **Metrics Exported**:
  - `mcp.tool.calls` (Counter) - Number of tool calls with server/tool attribution
  - `mcp.tool.duration` (Histogram) - Tool execution time in milliseconds
  - `mcp.tool.errors` (Counter) - Failed tool calls with error attribution

### Attributes

All telemetry includes server lineage preservation:
- `mcp.server.name` - Name of the MCP server
- `mcp.server.type` - Type (docker, sse, stdio, unknown)
- `mcp.tool.name` - Name of the tool being called
- `mcp.server.image` - Docker image (for containerized servers)
- `mcp.server.endpoint` - SSE/Remote endpoint (when applicable)

## Configuration

### Default Configuration

By default, MCP Gateway uses the Docker context's OTEL configuration:

```bash
# View current Docker context OTEL settings
docker context inspect | jq '.[].Metadata.otel'
```

Typical output:
```json
{
  "OTEL_EXPORTER_OTLP_ENDPOINT": "unix:///Users/username/.docker/run/user-analytics.otlp.grpc.sock"
}
```

### Environment Variable Override

You can override the endpoint for testing or custom deployments:

```bash
export DOCKER_CLI_OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317
docker mcp gateway run --catalog catalog.yaml
```

### Debug Mode

Enable debug logging to see telemetry initialization and tool call events:

```bash
export DOCKER_MCP_TELEMETRY_DEBUG=1
docker mcp gateway run --catalog catalog.yaml
```

Debug output appears on stderr and includes:
- Provider initialization status
- Tool call instrumentation events
- Metric creation confirmations

## Testing Telemetry

### Quick Test

Use the provided test script for a complete telemetry validation:

```bash
cd docs/telemetry/testing
./test-telemetry.sh --full
```

### Manual Testing

1. **Start an OTEL Collector**:
```bash
docker run --rm -d --name otel-debug \
  -p 4317:4317 -p 4318:4318 \
  -v docs/telemetry/testing/otel-collector-config.yaml:/config.yaml \
  otel/opentelemetry-collector:latest --config=/config.yaml
```

2. **Run MCP Gateway with Debug Mode**:
```bash
export DOCKER_MCP_TELEMETRY_DEBUG=1
export DOCKER_CLI_OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317
docker mcp gateway run --catalog ~/.docker/mcp/catalogs/docker-mcp.yaml \
  --transport sse --port 3000
```

3. **Make Tool Calls**:
Connect an MCP client to `http://localhost:3000/sse` and execute tools.

4. **View Telemetry**:
```bash
docker logs otel-debug | grep "mcp\."
```

### Expected Output

When telemetry is working correctly, you'll see:

**Traces**:
```
Span #0
    Trace ID       : 3482c4807337c24926eca041191e9d74
    Name           : mcp.tool.call
    Kind           : Client
    Status code    : Ok
Attributes:
     -> mcp.tool.name: Str(getPersonalNamespace)
     -> mcp.server.name: Str(dockerhub)
     -> mcp.server.type: Str(docker)
```

**Metrics**:
```
Metric #0
Descriptor:
     -> Name: mcp.tool.calls
     -> Description: Number of tool calls executed
NumberDataPoints #0
     -> mcp.server.name: Str(dockerhub)
     -> mcp.tool.name: Str(getPersonalNamespace)
Value: 2
```

## Troubleshooting

### No Telemetry Visible

1. **Check Provider Types**:
   ```bash
   DOCKER_MCP_TELEMETRY_DEBUG=1 docker mcp gateway run --dry-run
   ```
   Look for: `Tracer provider type: *trace.TracerProvider` (not noop)

2. **Verify Endpoint Configuration**:
   - Check Docker context: `docker context inspect`
   - Check environment: `echo $DOCKER_CLI_OTEL_EXPORTER_OTLP_ENDPOINT`

3. **Ensure Tool Calls Are Made**:
   - Telemetry is event-driven
   - No tool calls = no telemetry
   - Use SSE mode with an MCP client to trigger calls

### Debug Logging Not Appearing

- Ensure `DOCKER_MCP_TELEMETRY_DEBUG=1` is exported
- Check stderr output (debug logs go to stderr, not stdout)
- Verify the plugin was rebuilt after adding debug code

### Collector Connection Issues

- Verify collector is running: `docker ps | grep otel`
- Check port availability: `lsof -i :4317`
- Test connectivity: `telnet localhost 4317`

## Implementation Details

### Phase 1: Tool Call Telemetry (Completed)
- ✅ Basic telemetry package setup
- ✅ Tool handler instrumentation
- ✅ Server type inference
- ✅ Debug logging capability

### Phase 2: Extended Operations (Future)
- Prompt operation telemetry
- Resource operation telemetry
- Command-level spans

### Phase 3: Advanced Features (Future)
- Interceptor instrumentation
- Session-level tracking
- Connection pool metrics

## File Structure

```
docs/telemetry/
├── README.md                      # This file
└── testing/
    ├── otel-collector-config.yaml # Collector configuration for testing
    └── test-telemetry.sh          # Automated test script
```

## Development Guidelines

### Adding New Telemetry

1. **Use Existing Providers**:
   ```go
   tracer := otel.GetTracerProvider().Tracer("github.com/docker/mcp-gateway")
   meter := otel.GetMeterProvider().Meter("github.com/docker/mcp-gateway")
   ```

2. **Preserve Server Lineage**:
   Always include server attribution in spans and metrics.

3. **Non-Blocking Operations**:
   Telemetry must never block or fail operations.

4. **Debug Support**:
   Add debug logging behind `DOCKER_MCP_TELEMETRY_DEBUG` flag.

### Testing Changes

1. Build the plugin: `GOWORK=off make docker-mcp`
2. Run the test suite: `./docs/telemetry/testing/test-telemetry.sh --full`
3. Verify with real tool calls in SSE mode
4. Check both traces and metrics appear correctly

## References

- [OpenTelemetry Documentation](https://opentelemetry.io/docs/)
- [Docker CLI Telemetry](https://github.com/docker/cli/tree/master/cli/command)
- [MCP Protocol Specification](https://modelcontextprotocol.io/)