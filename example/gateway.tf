// A gateway configuration receives telemetry from edge collectors with
// OTLP and forwards it to an OTLP backend. The role=gateway label marks it
// as a gateway configuration. Requires Bindplane v1.103.0 or newer.
resource "bindplane_configuration_v2" "gateway" {
  lifecycle {
    create_before_destroy = true
  }

  rollout  = true
  name     = "example-gateway-config"
  platform = "linux"

  labels = {
    role = "gateway"
  }

  source {
    name = bindplane_source.otlp-default.name
    route {
      route_id       = "gateway-logs"
      telemetry_type = "logs"
      components = [
        "destinations/${bindplane_destination.otlp.id}"
      ]
    }
    route {
      route_id       = "gateway-metrics"
      telemetry_type = "metrics"
      components = [
        "destinations/${bindplane_destination.otlp.id}"
      ]
    }
    route {
      route_id       = "gateway-traces"
      telemetry_type = "traces"
      components = [
        "destinations/${bindplane_destination.otlp.id}"
      ]
    }
  }

  destination {
    route_id = bindplane_destination.otlp.id
    name     = bindplane_destination.otlp.name
  }
}

// A group of gateway collectors behind a load balancer. Collectors join
// the gateway when they are installed with the label gateway=example-gateway.
resource "bindplane_gateway" "gateway" {
  name                   = "example-gateway"
  display_name           = "Example Gateway"
  description            = "Gateway collectors behind a load balancer"
  configuration          = bindplane_configuration_v2.gateway.name
  load_balancer_hostname = "gateway.corp.net"
}
