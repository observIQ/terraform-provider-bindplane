// Inline connectors, processor group parameters, and inline processors.
// test.sh sets TF_VAR_enable_inline based on BINDPLANE_VERSION.
variable "enable_inline" {
  type    = bool
  default = false
}

resource "bindplane_configuration_v2" "inline" {
  count = var.enable_inline ? 1 : 0

  lifecycle {
    create_before_destroy = true
  }

  rollout  = true
  name     = "my-inline-config"
  platform = "linux"

  source {
    name = bindplane_source.journald.name

    processor {
      type = "filter_severity"
      parameters_json = jsonencode([
        {
          name  = "severity"
          value = "INFO"
        }
      ])
    }

    route {
      route_id       = "router"
      telemetry_type = "logs"
      components = [
        "connectors/router"
      ]
    }
  }

  connector {
    route_id = "router"
    type     = "routing"
    parameters_json = jsonencode([
      {
        name  = "telemetry_types"
        value = ["Logs"]
      },
      {
        name = "routes"
        value = [
          {
            id = "production"
            condition = {
              ottl = "attributes[\"env\"] == \"production\""
            }
          },
          {
            id = "default"
          }
        ]
      }
    ])

    route {
      route_id       = "production"
      telemetry_type = "logs"
      components = [
        "processors/production"
      ]
    }

    route {
      route_id       = "default"
      telemetry_type = "logs"
      components = [
        "destinations/${bindplane_destination.google.id}"
      ]
    }
  }

  processor_group {
    route_id = "production"
    parameters_json = jsonencode([
      {
        name  = "telemetry_types"
        value = ["Logs"]
      }
    ])

    processor {
      name = bindplane_processor.batch.name
    }

    processor {
      type         = "batch"
      display_name = "Inline Batch"
      parameters_json = jsonencode([
        {
          name  = "send_batch_size"
          value = 200
        }
      ])
    }

    route {
      route_id       = "destination"
      telemetry_type = "logs"
      components = [
        "destinations/${bindplane_destination.google.id}"
      ]
    }
  }

  destination {
    route_id = bindplane_destination.google.id
    name     = bindplane_destination.google.name
  }
}
