package catalog

import (
	"context"
	"encoding/json"
	"fmt"
)

func Ls(ctx context.Context, outputJSON bool) error {
	cfg, err := ReadConfigWithDefaultCatalog(ctx)
	if err != nil {
		return err
	}

	if outputJSON {
		data, err := json.Marshal(cfg)
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	} else {
		humanPrintCatalog(*cfg)
	}

	return nil
}

func humanPrintCatalog(cfg Config) {
	if len(cfg.Catalogs) == 0 {
		fmt.Println("No catalogs configured.")
		return
	}

	for name, catalog := range cfg.Catalogs {
		fmt.Printf("%s: %s\n", name, catalog.DisplayName)
	}
}
