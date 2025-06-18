package backup

import (
	"context"
	"encoding/json"

	"github.com/docker/docker-mcp/cmd/docker-mcp/internal/config"
	"github.com/docker/docker-mcp/cmd/docker-mcp/internal/desktop"
)

func Restore(ctx context.Context, backupData []byte) error {
	var backup Backup
	if err := json.Unmarshal(backupData, &backup); err != nil {
		return err
	}

	if err := config.WriteConfig([]byte(backup.Config)); err != nil {
		return err
	}
	if err := config.WriteRegistry([]byte(backup.Registry)); err != nil {
		return err
	}
	if err := config.WriteCatalog([]byte(backup.Catalog)); err != nil {
		return err
	}

	for name, content := range backup.CatalogFiles {
		if err := config.WriteCatalogFile(name, []byte(content)); err != nil {
			return err
		}
	}

	secretsClient := desktop.NewSecretsClient()
	for _, secret := range backup.Secrets {
		if err := secretsClient.SetJfsSecret(ctx, desktop.Secret{
			Name:     secret.Name,
			Value:    secret.Value,
			Provider: secret.Provider,
		}); err != nil {
			return err
		}
	}

	if err := secretsClient.SetJfsPolicy(ctx, backup.Policy); err != nil {
		return err
	}

	return nil
}
