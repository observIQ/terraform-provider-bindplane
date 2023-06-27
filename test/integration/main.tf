terraform {
  required_providers {
    bindplane = {
      source = "observiq/bindplane"
    }
  }
}

provider "bindplane" {
  remote_url = "https://localhost:3100"
  username = "tfu"
  password = "tfp"

  // server's certificate is signed by this CA
  tls_certificate_authority = "../../internal/client/tls/bindplane-ca.crt"

  // mtls client auth
  tls_certificate = "../../internal/client/tls/bindplane-client.crt"
  tls_private_key = "../../internal/client/tls/bindplane-client.key"

  // invalid mtls, client ca is not trusted by the server
  // tls_certificate = "../../internal/client/tls/test-client.crt"
  // tls_private_key = "../../internal/client/tls/test-client.key"
}

resource "bindplane_configuration" "config" {
  name = "testtf"
  labels = {
    purpose = "tf"
  }
  match_labels = {
    purpose = "tf"
  }
  raw_configuration = <<EOT
receivers:
  prometheus:
    config:
      scrape_configs:
        - job_name: 'collector'
          scrape_interval: 10s
          static_configs:
            - targets:
                - 'localhost:8888'
exporters:
  logging:
service:
  pipelines:
    metrics:
      receivers:
        - prometheus
      exporters:
        - logging
EOT
}

resource "bindplane_destination" "google_dest" {
  name = "google-test"
  type = "googlecloud"
  parameters = {
    "project": "abcd"
  }
}
