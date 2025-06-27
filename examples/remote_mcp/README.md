# Using the MCP Gateway with Docker Compose

This is a simple example of running the MCP Gateway that proxies to other remote MCP servers.

## With Docker Compose

+ Doesn't rely on the MCP Toolkit UI or the Docker socket. Can run anywhere, even if Docker Desktop is not available.
+ Defines the list of enabled servers from the gateway's command line, with `--server`.
+ Uses a custom catalog that lists a single remote MCP server.
+ Uses SSE for the transport and can be connected to via `http://localhost:8811/sse`.

```console
docker compose up
```

Add client services, like Agents, that connect with the `sse` protocol on port `8811`.
If needed, the protocol can be changed to `stdio` or `streaming` with `--transport=stdio` or `--transport=streaming`.

## On the host

**Run the Gateway**:

```console
docker mcp gateway run --servers=gitmcpmoby --catalog=./catalog.yaml
```

**Test a tool call**:

```console
docker mcp tools --gateway-arg="--servers=gitmcpmoby,--catalog=./catalog.yaml" call fetch_moby_documentation
```
