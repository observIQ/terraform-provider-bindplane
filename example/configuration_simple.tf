resource "bindplane_configuration" "configuration-simple" {
  // When removing a component from a configuration and deleting that
  // component during the same apply, we want to update the configuration
  // before the component is deleted.
  lifecycle {
    create_before_destroy = true
  }

  rollout  = false
  name     = "example-configuration-simple"
  platform = "linux"

  source {
    name = bindplane_source.host-custom.name
  }

  destination {
    name = bindplane_destination.grafana.name
  }
}
