// Gateways require Bindplane v1.103.0 or newer. test.sh sets
// TF_VAR_enable_gateway based on BINDPLANE_VERSION.
variable "enable_gateway" {
  type    = bool
  default = false
}

resource "bindplane_configuration_v2" "gateway" {
  count = var.enable_gateway ? 1 : 0

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
        "destinations/${bindplane_destination.prometheus.id}"
      ]
    }
  }

  destination {
    route_id = bindplane_destination.prometheus.id
    name     = bindplane_destination.prometheus.name
  }
}

resource "bindplane_gateway" "gateway" {
  count = var.enable_gateway ? 1 : 0

  name                   = "my-gateway"
  display_name           = "My Gateway"
  description            = "Gateway collectors behind a load balancer"
  configuration          = bindplane_configuration_v2.gateway[0].name
  load_balancer_hostname = "gateway.example.com"
}
