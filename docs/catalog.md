# MCP Catalog

## Testing MCP Servers

Sometimes, it's useful to run the MCP Gateway with your own MCP Server.

The standard way to submit new MCP Servers to Docker's catalog is
to go through this process: https://github.com/docker/mcp-registry?tab=readme-ov-file#-official-docker-mcp-registry

However, before you do that, you might want to craft a custom catalog
with some more entries.

### Custom catalog

The way to do it is to create a new `/path/to/custom_catalog.yaml` file
that contains every server you need and then run:

```console
docker mcp catalog import /path/to/custom_catalog.yaml
```

### Reset

When you're done with testing, run:

```console
docker mcp catalog reset
```
