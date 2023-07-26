resource "bindplane_source" "udp-default" {
  rollout = true
  name = "example-udp-default"
  type = "udp"
}

resource "bindplane_source" "udp-custom" {
  rollout = true
  name = "example-udp-custom"
  type = "udp"
  parameters_json = jsonencode(
    [
      {
        "name": "listen_ip",
        "value": "0.0.0.0"
      },
      {
        "name": "listen_port",
        "value": 8001
      },
      {
        "name": "log_type",
        "value": "udp"
      },
      {
        "name": "parse_format",
        "value": "regex"
      },
      {
        "name": "regex_pattern",
        "value": "%Y-%m-%d %H:%M:%S"
      },
      {
        "name": "parse_timestamp",
        "value": true
      },
      {
        "name": "timestamp_field",
        "value": "timestamp"
      },
      {
        "name": "parse_timestamp_format",
        "value": "ISO8601"
      },
      {
        "name": "epoch_timestamp_format",
        "value": "s"
      },
      {
        "name": "manual_timestamp_format",
        "value": "%Y-%m-%dT%H:%M:%S.%f%z"
      },
      {
        "name": "timezone",
        "value": "America/Detroit"
      },
      {
        "name": "parse_severity",
        "value": true
      },
      {
        "name": "severity_field",
        "value": "severity"
      },
      {
        "name": "parse_to",
        "value": "body"
      }
    ]
  )
}
