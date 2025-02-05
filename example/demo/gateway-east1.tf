resource "bindplane_source" "bindplane_gateway" {
  rollout = true
  name    = "bindplane-gateway-tf"
  type    = "bindplane_gateway"
}

resource "bindplane_configuration" "gateway-east1" {
  // When removing a component from a configuration and deleting that
  // component during the same apply, we want to update the configuration
  // before the component is deleted.
  lifecycle {
    create_before_destroy = true
  }

  rollout  = true
  name     = "gateway-east1-tf"
  platform = "linux"

  source {
    name = bindplane_source.bindplane_gateway.name
  }

  destination {
    name = bindplane_destination.nop.name
  }
}
