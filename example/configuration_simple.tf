resource "bindplane_configuration" "configuration-simple" {
  rollout = false
  name = "example-configuration-simple"
  platform = "linux"

  source {
    name = bindplane_source.host-custom.name
  }

  destination {
    name = bindplane_destination.grafana.name
  }
}
