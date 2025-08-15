package interceptors

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/docker/mcp-gateway/cmd/docker-mcp/internal/telemetry"
)

// TelemetryMiddleware tracks ListTools calls and other gateway operations
func TelemetryMiddleware() mcp.Middleware[*mcp.ServerSession] {
	return func(next mcp.MethodHandler[*mcp.ServerSession]) mcp.MethodHandler[*mcp.ServerSession] {
		return func(ctx context.Context, session *mcp.ServerSession, method string, params mcp.Params) (mcp.Result, error) {
			// Track ListTools calls
			if method == "tools/list" {
				telemetry.RecordListTools(ctx)
			}
			
			// Call the next handler
			return next(ctx, session, method, params)
		}
	}
}