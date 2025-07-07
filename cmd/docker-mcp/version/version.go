package version

var Version = "HEAD"

func UserAgent() string {
	return "docker/mcp_gateway/v/" + Version
}
