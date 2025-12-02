package oauth_test

import (
	"os"
	"testing"

	"github.com/docker/mcp-gateway/pkg/oauth"
)

func TestIsCEMode_ExplicitOverride(t *testing.T) {
	// Test explicit true
	t.Run("explicit_true", func(t *testing.T) {
		os.Setenv("DOCKER_MCP_USE_CE", "true")
		defer os.Unsetenv("DOCKER_MCP_USE_CE")

		if !oauth.IsCEMode() {
			t.Error("Expected CE mode when DOCKER_MCP_USE_CE=true")
		}
	})

	// Test explicit false
	t.Run("explicit_false", func(t *testing.T) {
		os.Setenv("DOCKER_MCP_USE_CE", "false")
		defer os.Unsetenv("DOCKER_MCP_USE_CE")

		if oauth.IsCEMode() {
			t.Error("Expected Desktop mode when DOCKER_MCP_USE_CE=false")
		}
	})

	// Test empty string (should fallback to auto-detect)
	t.Run("empty_string_fallback", func(t *testing.T) {
		os.Setenv("DOCKER_MCP_USE_CE", "")
		defer os.Unsetenv("DOCKER_MCP_USE_CE")

		// Should fallback to auto-detection
		// Result depends on whether Desktop socket exists
		_ = oauth.IsCEMode()
	})
}

func TestIsCEMode_AutoDetect(t *testing.T) {
	// Unset env var to test auto-detection
	os.Unsetenv("DOCKER_MCP_USE_CE")

	// Result depends on environment
	// In CI/Linux CE: should be true
	// With Docker Desktop: should be false
	result := oauth.IsCEMode()
	t.Logf("Auto-detected CE mode: %v", result)

	// Verify function returns consistently
	result2 := oauth.IsCEMode()
	if result != result2 {
		t.Error("IsCEMode() should return consistent results")
	}
}

func TestIsCEMode_OverrideTakesPrecedence(t *testing.T) {
	// Override should take precedence over auto-detection
	t.Run("override_true_precedence", func(t *testing.T) {
		os.Setenv("DOCKER_MCP_USE_CE", "true")
		defer os.Unsetenv("DOCKER_MCP_USE_CE")

		// Should return true regardless of socket existence
		if !oauth.IsCEMode() {
			t.Error("Expected CE mode when explicitly set to true")
		}
	})

	t.Run("override_false_precedence", func(t *testing.T) {
		os.Setenv("DOCKER_MCP_USE_CE", "false")
		defer os.Unsetenv("DOCKER_MCP_USE_CE")

		// Should return false regardless of socket existence
		if oauth.IsCEMode() {
			t.Error("Expected Desktop mode when explicitly set to false")
		}
	})
}
