package oauth

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/mcp-gateway/cmd/docker-mcp/internal/desktop"
)

func Authorize(ctx context.Context, app string, scopes string) error {
	client := desktop.NewAuthClient()

	authResponse, err := client.PostOAuthApp(ctx, app, scopes, false)
	if err != nil {
		// Check if error is about provider not found
		if strings.Contains(err.Error(), "not found") {
			return fmt.Errorf("OAuth provider does not exist")
		}
		return err
	}

	// Check if the response contains a valid browser URL
	if authResponse.BrowserURL == "" {
		return fmt.Errorf("OAuth provider does not exist")
	}

	fmt.Printf("Opening your browser for authentication. If it doesn't open automatically, please visit: %s\n", authResponse.BrowserURL)

	return nil
}
