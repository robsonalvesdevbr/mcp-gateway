package backup

import (
	"context"
	"encoding/json"

	"github.com/docker/docker-mcp/cmd/docker-mcp/internal/docker"
)

func Restore(_ context.Context, _ docker.Client, backupData []byte) error {
	var backup Backup
	if err := json.Unmarshal(backupData, &backup); err != nil {
		return err
	}

	return nil
}
