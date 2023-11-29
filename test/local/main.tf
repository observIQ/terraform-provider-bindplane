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

resource "bindplane_source" "host" {
  rollout = true
  name = "my-host"
  type = "host"
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

resource "bindplane_source" "journald" {
  rollout = true
  name = "my-journald"
  type = "journald"
}

resource "bindplane_destination" "custom" {
  rollout = true
  name = "example-custom"
  type = "custom"
  parameters_json = jsonencode(
    [
      {
        "name": "telemetry_types",
        "value": [
          "Metrics",
          "Logs",
          "Traces"
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
    name = bindplane_source.journald.name
    processors = [
      bindplane_processor.add_fields.name
    ]
  }

  source {
    name = bindplane_source.host.name
    processors = [
      bindplane_processor.add_fields.name
    ]
  }


  destination {
    name = bindplane_destination.custom.name
    processors = [
      bindplane_processor.batch.name,

      // order matters here
      bindplane_processor.include-flog.name,
      bindplane_processor.time-parse-http-datatime.name,
      bindplane_processor.promoted-cleanup.name
    ]
  }
}

resource "bindplane_processor" "json-parse-body" {
  rollout = false
  name = "json-parse-body"
  type = "parse_json"
  parameters_json = jsonencode(
    [
      {
        "name": "enable_logs",
        "value": true
      },
      {
        "name": "log_condition",
        "value": "true"
      },
      {
        "name": "log_source_field_type",
        "value": "Body"
      },
      {
        "name": "log_body_source_field",
        "value": ""
      },
      {
        "name": "log_target_field_type",
        "value": "Body"
      }
    ]
  )
}

resource "bindplane_processor" "promoted-cleanup" {
  rollout = true
  name = "promoted-cleanup"
  type = "delete_fields"
  parameters_json = jsonencode(
    [
      {
        "name": "enable_logs",
        "value": true
      },
      {
        "name": "log_body_keys",
        "value": [
          "datetime"
        ]
      }
    ]
  )
}

resource "bindplane_processor" "include-flog" {
  rollout = false
  name = "include-flog"
  type = "filter_field"
  parameters_json = jsonencode(
    [
      {
        "name": "telemetry_types",
        "value": [
          "Metrics",
          "Logs",
          "Traces"
        ]
      },
      {
        "name": "action",
        "value": "include"
      },
      {
        "name": "match_type",
        "value": "regexp"
      },
      {
        "name": "bodies",
        "value": {
          "status": ".*"
        }
      }
    ]
  )
}

resource "bindplane_processor" "time-parse-http-datatime" {
  rollout = false
  name = "time-parse-http-datatime"
  type = "parse_timestamp"
  parameters_json = jsonencode(
    [
      {
        "name": "enable_logs",
        "value": true
      },
      {
        "name": "log_condition",
        "value": "body[\"datetime\"] != nil"
      },
      {
        "name": "log_field_type",
        "value": "Body"
      },
      {
        "name": "log_source_field",
        "value": "datetime"
      },
      {
        "name": "log_time_format",
        "value": "Manual"
      },
      {
        "name": "log_epoch_layout",
        "value": "s"
      },
      {
        "name": "log_manual_timestamp_layout",
        "value": "%d/%b/%Y:%H:%M:%S %z"
      }
    ]
  )
}
