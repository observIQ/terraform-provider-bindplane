terraform {
  required_providers {
    bindplane = {
      source = "observiq/bindplane"
    }
  }
}

provider "bindplane" {
  remote_url      = "https://localhost:3001"
  username        = "admin"
  password        = "password"
  tls_skip_verify = true
}

resource "bindplane_source" "host" {
  rollout = true
  name    = "my-host"
  type    = "host"
  parameters_json = jsonencode(
    [
      {
        "name": "collection_interval",
        "value": 32
      },
      {
        "name": "enable_process",
        "value": false
      }
    ]
  )
}

resource "bindplane_destination" "custom" {
  rollout = true
  name    = "example-custom"
  type    = "custom"
  parameters_json = jsonencode(
    [
      {
        "name": "telemetry_types",
        "value": [
          "Logs",
        ]
      },
      {
        "name": "configuration",
        "value": <<EOT
logging:
  verbosity: detailed
  sampling_initial: 5
  sampling_thereafter: 200
EOT
      }
    ]
  )
}

resource "bindplane_configuration" "configuration" {
  lifecycle {
    create_before_destroy = true
  }

  rollout = true
  name     = "my-config"
  platform = "linux"

  source {
    name = bindplane_source.host.name
  }

  destination {
    name = bindplane_destination.custom.name
  }
}
