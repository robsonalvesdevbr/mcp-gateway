# docker mcp gateway run

<!---MARKER_GEN_START-->
Run the gateway

### Options

| Name                  | Type          | Default           | Description                                                                                                                                  |
|:----------------------|:--------------|:------------------|:---------------------------------------------------------------------------------------------------------------------------------------------|
| `--block-network`     | `bool`        |                   | Block tools from accessing forbidden network resources                                                                                       |
| `--block-secrets`     | `bool`        | `true`            | Block secrets from being/received sent to/from tools                                                                                         |
| `--catalog`           | `string`      | `docker-mcp.yaml` | path to the docker-mcp.yaml catalog (absolute or relative to ~/.docker/mcp/catalogs/)                                                        |
| `--config`            | `string`      | `config.yaml`     | path to the config.yaml (absolute or relative to ~/.docker/mcp/)                                                                             |
| `--cpus`              | `int`         | `1`               | CPUs allocated to each MCP Server (default is 1)                                                                                             |
| `--debug-dns`         | `bool`        |                   | Debug DNS resolution                                                                                                                         |
| `--dry-run`           | `bool`        |                   | Start the gateway but do not listen for connections (useful for testing the configuration)                                                   |
| `--interceptor`       | `stringArray` |                   | List of interceptors to use (format: when:type:path, e.g. 'before:exec:/bin/path')                                                           |
| `--log-calls`         | `bool`        | `true`            | Log calls to the tools                                                                                                                       |
| `--long-lived`        | `bool`        |                   | Containers are long-lived and will not be removed until the gateway is stopped, useful for stateful servers                                  |
| `--memory`            | `string`      | `2Gb`             | Memory allocated to each MCP Server (default is 2Gb)                                                                                         |
| `--port`              | `int`         | `0`               | TCP port to listen on (default is to listen on stdio)                                                                                        |
| `--registry`          | `string`      | `registry.yaml`   | path to the registry.yaml (absolute or relative to ~/.docker/mcp/)                                                                           |
| `--secrets`           | `string`      | `docker-desktop`  | colon separated paths to search for secrets. Can be `docker-desktop` or a path to a .env file (default to using Docker Deskop's secrets API) |
| `--servers`           | `stringSlice` |                   | names of the servers to enable (if non empty, ignore --registry flag)                                                                        |
| `--static`            | `bool`        |                   | Enable static mode (aka pre-started servers)                                                                                                 |
| `--tools`             | `stringSlice` |                   | List of tools to enable                                                                                                                      |
| `--transport`         | `string`      | `stdio`           | stdio, sse or streaming (default is stdio)                                                                                                   |
| `--verbose`           | `bool`        |                   | Verbose output                                                                                                                               |
| `--verify-signatures` | `bool`        |                   | Verify signatures of the server images                                                                                                       |
| `--watch`             | `bool`        | `true`            | Watch for changes and reconfigure the gateway                                                                                                |


<!---MARKER_GEN_END-->

