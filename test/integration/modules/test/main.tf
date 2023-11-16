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

resource "bindplane_source" "host" {
  rollout = true
  name = "my-host"
  type = "host"
  parameters_json = jsonencode(
    [
      {
        "name": "collection_interval",
        "value": 20
      },
      {
        "name": "enable_process",
        "value": false
      },
      {
        "name": "metric_filtering",
        "value": [
          "system.disk.operation_time"
        ]
      }
    ]
  )
}

resource "bindplane_processor" "batch" {
  rollout = true
  name = "my-batch"
  type = "batch"
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
