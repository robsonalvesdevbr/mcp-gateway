package policy

import (
	"context"

	"github.com/docker/mcp-gateway/cmd/docker-mcp/internal/desktop"
)

func Set(ctx context.Context, data string) error {
	return desktop.NewSecretsClient().SetJfsPolicy(ctx, data)
}
