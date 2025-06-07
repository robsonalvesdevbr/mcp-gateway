package interceptors

import (
	"context"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func LogCalls(next server.ToolHandlerFunc) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		start := time.Now()

		tool := request.Params.Name
		arguments := argumentsToString(request.Params.Arguments)

		logf("- Calling tool %s with arguments: %s\n", tool, arguments)

		result, err := next(ctx, request)
		if err != nil {
			return result, err
		}

		logf("> Calling tool %s took: %s\n", tool, time.Since(start))

		return result, nil
	}
}
