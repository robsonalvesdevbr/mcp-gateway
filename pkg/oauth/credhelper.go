package oauth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"

	"github.com/docker/docker-credential-helpers/client"
	"github.com/docker/docker-credential-helpers/credentials"

	"github.com/docker/mcp-gateway/pkg/desktop"
)

// CredentialHelper provides secure access to OAuth tokens via docker-credential-desktop
type CredentialHelper struct {
	credentialHelper credentials.Helper
}

// NewOAuthCredentialHelper creates a new OAuth credential helper
func NewOAuthCredentialHelper() *CredentialHelper {
	return &CredentialHelper{
		credentialHelper: newOAuthHelper(),
	}
}

// GetOAuthToken retrieves an OAuth token for the specified server
// It follows this flow:
// 1. Get DCR client info to retrieve provider name and authorization endpoint
// 2. Construct credential key using: [AuthorizationEndpoint]/[ProviderName]
// 3. Retrieve token from docker-credential-desktop
func (h *CredentialHelper) GetOAuthToken(ctx context.Context, serverName string) (string, error) {
	// Step 1: Get DCR client info (includes stored provider name)
	client := desktop.NewAuthClient()
	dcrClient, err := client.GetDCRClient(ctx, serverName)
	if err != nil {
		return "", fmt.Errorf("no DCR client found for %s: %w", serverName, err)
	}

	// Step 2: Construct credential key using authorization endpoint + provider name
	credentialKey := fmt.Sprintf("%s/%s", dcrClient.AuthorizationEndpoint, dcrClient.ProviderName)

	// Step 3: Retrieve token from docker-credential-desktop
	_, tokenSecret, err := h.credentialHelper.Get(credentialKey)
	if err != nil {
		if credentials.IsErrCredentialsNotFound(err) {
			return "", fmt.Errorf("OAuth token not found for %s (key: %s). Run 'docker mcp oauth authorize %s' to authenticate", serverName, credentialKey, serverName)
		}
		return "", fmt.Errorf("failed to retrieve OAuth token for %s: %w", serverName, err)
	}

	if tokenSecret == "" {
		return "", fmt.Errorf("empty OAuth token found for %s", serverName)
	}

	// The secret is base64-encoded JSON, decode it first
	tokenJSON, err := base64.StdEncoding.DecodeString(tokenSecret)
	if err != nil {
		return "", fmt.Errorf("failed to decode OAuth token for %s: %w", serverName, err)
	}

	// Parse the JSON to extract the actual access token
	var tokenData struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
	}
	if err := json.Unmarshal(tokenJSON, &tokenData); err != nil {
		return "", fmt.Errorf("failed to parse OAuth token JSON for %s: %w", serverName, err)
	}

	if tokenData.AccessToken == "" {
		return "", fmt.Errorf("empty OAuth access token found for %s", serverName)
	}

	return tokenData.AccessToken, nil
}

// newOAuthHelper creates a credential helper for OAuth token access
func newOAuthHelper() credentials.Helper {
	return oauthHelper{
		program: newShellProgramFunc("docker-credential-desktop"),
	}
}

// newShellProgramFunc creates programs that are executed in a Shell.
func newShellProgramFunc(name string) client.ProgramFunc {
	return func(args ...string) client.Program {
		return &shell{cmd: exec.CommandContext(context.Background(), name, args...)}
	}
}

// shell invokes shell commands to talk with a remote credentials-helper.
type shell struct {
	cmd *exec.Cmd
}

// Output returns responses from the remote credentials-helper.
func (s *shell) Output() ([]byte, error) {
	return s.cmd.Output()
}

// Input sets the input to send to a remote credentials-helper.
func (s *shell) Input(in io.Reader) {
	s.cmd.Stdin = in
}

// oauthHelper wraps credential helper program for OAuth token access.
type oauthHelper struct {
	program client.ProgramFunc
}

func (h oauthHelper) List() (map[string]string, error) {
	return map[string]string{}, nil
}

// Add stores new credentials (not used for OAuth token retrieval)
func (h oauthHelper) Add(_ *credentials.Credentials) error {
	return fmt.Errorf("OAuth credential helper is read-only")
}

// Delete removes credentials (not used for OAuth token retrieval)
func (h oauthHelper) Delete(_ string) error {
	return fmt.Errorf("OAuth credential helper is read-only")
}

// Get returns the OAuth token for a given credential key
func (h oauthHelper) Get(credentialKey string) (string, string, error) {
	creds, err := client.Get(h.program, credentialKey)
	if err != nil {
		return "", "", err
	}
	return creds.Username, creds.Secret, nil
}

var _ credentials.Helper = oauthHelper{}
