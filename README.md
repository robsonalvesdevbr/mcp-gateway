# Docker MCP CLI
![build](https://github.com/docker/mcp-cli/actions/workflows/ci.yml/badge.svg)

## Docker MCP CLI Plugin
For local development, install the plugin via
```shell
make docker-mcp
```

Formatting and linting prerequisites:

- Go to the [GH PAT settings page](https://github.com/settings/tokens) and generate a new token.
- Select at least `read:packages` & `repo`.
- Add your PAT to your shell settings (e.g.: `~/.zshrc` or `~/.profile`) with the var `GIT_GHA_TOKEN`
- Restart your terminal

You should be able to run:
```shell
make format
make lint
```