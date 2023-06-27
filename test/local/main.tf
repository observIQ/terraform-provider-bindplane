terraform {
  required_providers {
    bindplane = {
      source = "observiq/bindplane"
    }
  }
}

provider "bindplane" {
  remote_url = "http://localhost:3001"
  username = "admin"
  password = "admin"
}

resource "bindplane_configuration" "raw" {
  name = "testtf-raw"
  labels = {
    purpose = "tf-raw"
  }
  match_labels = {
    purpose = "tf-raw"
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

resource "bindplane_configuration" "config" {
  name = "testtf"
  labels = {
    purpose = "tf"
  }
  match_labels = {
    purpose = "tf"
  }
  
  source {
    type = "host"
    parameters_json = jsonencode({
      "metric_filtering": [
        "system.disk.operation_time"
      ]
      "enable_process": false,
      "collection_interval": 20
    })
  }

  source {
    type = "otlp"
  }

  source {
    type = "otlp"
    parameters_json = jsonencode({
      "http_port": 44318,
      "grpc_port": 0
    })
  }

  destinations = [
    bindplane_destination.google_dest.name
  ]
}

resource "bindplane_destination" "google_dest" {
  name = "google-test"
  type = "googlecloud"
  parameters_json = jsonencode({
    "project": "abcd"
  })
}
