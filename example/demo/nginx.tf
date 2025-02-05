resource "bindplane_source" "nginx" {
  rollout = true
  name    = "nginx-tf"
  type    = "nginx"
}

resource "bindplane_configuration" "nginx" {
  // When removing a component from a configuration and deleting that
  // component during the same apply, we want to update the configuration
  // before the component is deleted.
  lifecycle {
    create_before_destroy = true
  }

  rollout  = true
  name     = "nginx-tf"
  platform = "linux"

  source {
    name = bindplane_source.nginx.name
  }

  destination {
    name = bindplane_destination.gateway-east1.name
  }
}
