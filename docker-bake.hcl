
group default {
  targets = [
    "all",
  ]
}

group all {
  targets = [
    "agents_gateway",
    "jcat",
    "l4proxy",
    "l7proxy",
  ]
}

# Required by docker/metadata-action and docker/bake-action gh actions.
target "docker-metadata-action" {}

target _base {
  inherits = ["docker-metadata-action"]
  output = ["type=docker"]
  platforms = ["linux/arm64", "linux/amd64"]
  attest = [
    {
      type = "provenance",
      mode = "max",
    },
    {
      type = "sbom",
    }
  ]
}

target jcat {
  inherits = ["_base"]
  context = "tools/jcat"
  output = ["type=image,name=docker/jcat"]
}


target l4proxy {
  inherits = ["_base"]
  context = "tools/l4proxy"
  output = ["type=image,name=docker/mcp-l4proxy:v1"]
}

target l7proxy {
  inherits = ["_base"]
  context = "tools/http_proxy"
  output = ["type=image,name=docker/mcp-l7proxy:v1"]
}

target agents_gateway {
  inherits = ["_base"]
  context = "."
  target = "agents_gateway"
  output = ["type=image,name=docker/agents_gateway:v2"]
}
