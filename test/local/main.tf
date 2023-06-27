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

resource "bindplane_raw_configuration" "raw" {
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
  rollout = true
  name = "testtf"
  labels = {
    purpose = "tf"
  }
  match_labels = {
    purpose = "tf"
  }

  destinations = [
    bindplane_destination.google_dest.name
  ]

  sources = [
    bindplane_source.otlp.name,
    bindplane_source.otlp-custom.name,
    bindplane_source.host.name
  ]
}

resource "bindplane_destination" "google_dest" {
  rollout = true
  name = "google-test"
  type = "googlecloud"
  parameters_json = jsonencode({
    "project": "abcd"
  })
}

resource "bindplane_source" "otlp" {
  rollout = true
  name = "otlp-default"
  type = "otlp"
}

resource "bindplane_source" "otlp-custom" {
  rollout = true
  name = "otlp-custom"
  type = "otlp"
  parameters_json = jsonencode({
    "http_port": 44313,
    "grpc_port": 0
  })
}

resource "bindplane_source" "host" {
  rollout = true
  name = "my-host"
  type = "host"
  parameters_json = jsonencode({
    "metric_filtering": [
      "system.disk.operation_time"
    ],
    "enable_process": false,
    "collection_interval": 20
  })
}
