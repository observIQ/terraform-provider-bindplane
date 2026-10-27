resource "bindplane_configuration_v2" "gateway" {
  lifecycle {
    create_before_destroy = true
  }

  rollout  = true
  name     = "my-gateway-config"
  platform = "linux"

  // Marks the configuration as a gateway configuration.
  labels = {
    role = "gateway"
  }

  source {
    name = bindplane_source.otlp.name
    route {
      route_id       = "gateway-logs"
      telemetry_type = "logs"
      components = [
        "destinations/${bindplane_destination.custom.id}"
      ]
    }
  }

  destination {
    route_id = bindplane_destination.custom.id
    name     = bindplane_destination.custom.name
  }
}

resource "bindplane_gateway" "gateway" {
  name                   = "my-gateway"
  display_name           = "My Gateway"
  description            = "Gateway collectors behind a load balancer"
  configuration          = bindplane_configuration_v2.gateway.name
  load_balancer_hostname = "gateway.example.com"
}
