# Using the MCP Gateway with Docker Compose and the MCP Toolkit

This is a more complex example of running the MCP Gateway with Docker Compose:

+ It does rely on the MCP Toolkit UI runnning in Docker Desktop.
+ All the configuration of the MCP servers is delegated to the MCP Toolkit.

## How to run

```console
docker compose up
```

Add client services, like Agents, that connect with the `stdio` protocol on port `8811`.
If needed, the protocol can be changed to `sse` or `streaming` with `--transport=sse` or `--transport=streaming`.
