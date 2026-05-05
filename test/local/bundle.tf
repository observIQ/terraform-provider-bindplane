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

resource "bindplane_source" "splunk_tcp" {
  rollout = false
  name    = "Splunk"
  type    = "splunk_tcp"
}

resource "bindplane_source" "fluent" {
  rollout = false
  name    = "Fluent"
  type    = "fluentforward"
}

resource "bindplane_processor" "json_parser" {
  rollout = false
  name    = "Parse-JSON-Body"
  type    = "parse_json"
}

resource "bindplane_processor" "severity_parser_v2" {
  rollout = false
  name = "Parse-Severity-HTTP-Status-v2"
  type = "parse_severity_v2"
  parameters_json = jsonencode(
  [
    {
      "name": "match",
      "value": "Body"
    },
    {
      "name": "body_severity_field",
      "value": "level"
    },
  ]
  )
}

resource "bindplane_processor" "time_parser" {
  rollout = false
  name    = "Parse-Timestamp"
  type    = "parse_timestamp"
  parameters_json = jsonencode(
    [
      {
        "name": "telemetry_types",
        "value": [
          "Logs"
        ]
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
        "name": "log_manual_timestamp_layout",
        "value": "%d/%b/%Y:%H:%M:%S %z"
      },
    ]
  )
}

resource "bindplane_processor" "cleanup_promoted" {
  rollout = false
  name    = "cleanup"
  type    = "delete_fields_v2"
  parameters_json = jsonencode(
    [
      {
        "name": "telemetry_types",
        "value": ["Logs"]
      },
      {
        "name": "resource_attributes",
        "value": ["level","datetime"]
      },
    ]
  )
}

resource "bindplane_processor_bundle" "parser" {
  rollout = true
  name = "generic-parse-bundle"

  processor {
    name = bindplane_processor.json_parser.name
  }

  processor {
    name = bindplane_processor.severity_parser_v2.name
  }

  processor {
    name = bindplane_processor.time_parser.name
  }

  processor {
    name = bindplane_processor.cleanup_promoted.name
  }
}

resource "bindplane_configuration" "splunk" {
  lifecycle {
    create_before_destroy = true
  }

  rollout  = true
  name     = "splunk"
  platform = "linux"

  source {
    name = bindplane_source.splunk_tcp.name
    processors = [
      bindplane_processor_bundle.parser.name,
    ]
  }

  destination {
    name = bindplane_destination.logging.name
  }
}

resource "bindplane_configuration" "fluent" {
  lifecycle {
    create_before_destroy = true
  }

  rollout  = true
  name     = "fluent"
  platform = "linux"

  source {
    name = bindplane_source.fluent.name
    processors = [
      bindplane_processor_bundle.parser.name,
    ]
  }

  destination {
    name = bindplane_destination.logging.name
  }
}
