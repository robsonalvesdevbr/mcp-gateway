package catalog

import (
	"context"
)

func Init(ctx context.Context) error {
	return Import(ctx, DockerCatalogName)
}
