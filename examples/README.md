# Using the MCP Gateway with Docker Compose and the MCP Toolkit

+ `minimal` - Simplest Compose file. Just one MCP Server, without configuration or secrets.
+ `client` - A Python client connecting to the MCP Gateway over http streaming transport.
+ `secrets` - Just one MCP Server, with a secret handled in an `.env` file.
+ `remote_mcp` - Uses the gateway as a proxy to a remote MCP server.
+ `mcp_toolkit` - Connect to the MCP Toolkit and let it handle all the configuration and secrets.
