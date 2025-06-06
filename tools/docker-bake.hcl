
group default {
  targets = [
    "all",
  ]
}

group all {
  targets = [
    "jcat",
  ]
}

# Required by docker/metadata-action and docker/bake-action gh actions.
target "docker-metadata-action" {}

target _base {
  inherits = ["docker-metadata-action"]
  output = ["type=docker"]
}

target jcat {
  inherits = ["_base"]
  context = "jcat"
  output = ["type=image,name=docker/jcat"]
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
