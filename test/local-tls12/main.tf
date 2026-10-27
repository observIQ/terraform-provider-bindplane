// Connects to Bindplane through a TLS 1.2-only proxy (nginx.conf).
// With the default tls_min_version (1.3) the provider cannot connect.
terraform {
  required_providers {
    bindplane = {
      source = "observiq/bindplane"
    }
  }
}

variable "tls_min_version" {
  type    = string
  default = null
}

variable "tls_certificate_authority" {
  type    = string
  default = "/tmp/bindplane.crt"
}

provider "bindplane" {
  remote_url                = "https://localhost:3002"
  username                  = "admin"
  password                  = "password"
  tls_certificate_authority = var.tls_certificate_authority
  tls_min_version           = var.tls_min_version
}

resource "bindplane_destination" "custom" {
  rollout = true
  name    = "tls12-custom"
  type    = "custom"
  parameters_json = jsonencode(
    [
      {
        "name" : "telemetry_types",
        "value" : ["Logs"]
      },
      {
        "name" : "configuration",
        "value" : "logging:\n  verbosity: basic\n"
      },
      {
        "name" : "drop_raw_copy",
        "value" : true
      }
    ]
  )
}
