terraform {
  required_providers {
    bindplane = {
      source = "observiq/bindplane"
    }
  }
}

provider "bindplane" {
  profile = "local"
  # remote_url = "http://localhost:3001"
  # username = "admin"
  # password = "admin"
}

resource "bindplane_raw_configuration" "raw" {
  rollout = true
  name = "testtf-raw"
  platform = "linux"
  labels = {
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
  platform = "linux"
  labels = {
    purpose = "tf"
  }

  destination {
    name = bindplane_destination.logging.name
    processors = [
      bindplane_processor.batch-options.name
    ]
  }

  destination {
    name = bindplane_destination.logging2.name
  }

  source {
    name = bindplane_source.otlp.name
    processors = [
      bindplane_processor.add_fields.name
    ]
  }

  source {
    name = bindplane_source.otlp-custom.name
  }

  source {
    name = bindplane_source.host.name
  }
}

// Do not attach to test config. Will fail to startup
// due to missing credentials. Used as an example
// on how to create a GCP destination.
resource "bindplane_destination" "google_dest" {
  rollout = true
  name = "google-test"
  type = "googlecloud"
  parameters_json = jsonencode(
    [
      {
        "name": "project",
        "value": "abcd"
      },
    ]
  )
}

resource "bindplane_destination" "logging" {
  rollout = true
  name = "logging"
  type = "custom"
  parameters_json = jsonencode(
    [
      {
        "name": "telemetry_types",
        "value": ["Metrics", "Logs", "Traces"]
      },
      {
        "name": "configuration",
        "value": "logging:"
      }
    ]
  )
}

resource "bindplane_destination" "logging2" {
  rollout = true
  name = "logging-2"
  type = "custom"
  parameters_json = jsonencode(
    [
      {
        "name": "telemetry_types",
        "value": ["Metrics", "Logs", "Traces"]
      },
      {
        "name": "configuration",
        "value": "logging:"
      }
    ]
  )
}

resource "bindplane_destination" "prometheus" {
  rollout = true
  name = "prometheus"
  type = "prometheus"
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
  parameters_json = jsonencode(
    [
      {
        "name": "http_port",
        "value": 44314
      },
      {
        "name": "grpc_port",
        "value": 0
      }
    ]
  )
}

resource "bindplane_processor" "batch-options" {
  rollout = true
  name = "my-batch-options"
  type = "batch"
  parameters_json = jsonencode(
    [
      {
        "name": "send_batch_size",
        "value": 200
      },
      {
        "name": "send_batch_max_size",
        "value": 400
      },
      {
        "name": "timeout",
        "value": "2s"
      }
    ]
  )
}


resource "bindplane_destination" "prometheus2" {
  rollout = true
  name = "my-prometheus"
  type = "prometheus"
  parameters_json = jsonencode(
    [
      {
        "name": "listen_address",
        "value": "0.0.0.0"
      },
      {
        "name": "listen_port",
        "value": 9000,
      },
      {
        "name": "namespace",
        "value": "otel"
      }
    ]
  )
}

resource "bindplane_processor" "batch2" {
  rollout = true
  name = "my-batch-2"
  type = "batch"
  parameters_json = jsonencode(
    [
      {
        "name": "send_batch_size",
        "value": 200
      },
      {
        "name": "send_batch_max_size",
        "value": 400
      },
      {
        "name": "timeout",
        "value": "2s"
      }
    ]
  )
}

resource "bindplane_source" "host" {
  rollout = true
  name = "my-host"
  type = "host"
  parameters_json = jsonencode(
    [
      {
        "name": "collection_interval",
        "value": 30
      },
      {
        "name": "enable_process",
        "value": false
      }
    ]
  )
}

resource "bindplane_source" "journald" {
  rollout = true
  name = "my-journald"
  type = "journald"
}

resource "bindplane_destination" "google" {
  rollout = true
  name = "my-google"
  type = "googlecloud"
}

resource "bindplane_processor" "add_fields" {
  rollout = true
  name = "add-fields"
  type = "add_fields"
  parameters_json = jsonencode(
    [
      {
        "name": "enable_logs"
        "value": true
      },
      {
        "name": "log_resource_attributes",
        "value": {
          "key": "value2"
        }
      }
    ]
  )
}

resource "bindplane_processor" "batch" {
  rollout = true
  name = "my-batch"
  type = "batch"
}

resource "bindplane_configuration" "configuration" {
  rollout = true
  name = "my-config"
  platform = "linux"
  labels = {
    environment = "production"
    managed-by  = "terraform"
  }

  source {
    name = bindplane_source.host.name
    processors = [
      bindplane_processor.add_fields.name
    ]
  }

  source {
    name = bindplane_source.journald.name
    processors = [
      bindplane_processor.add_fields.name
    ]
  }

  destination {
    name = bindplane_destination.google.name
    processors = [
      bindplane_processor.batch.name
    ]
  }
}
