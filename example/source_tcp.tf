resource "bindplane_source" "tcp-default" {
  rollout = true
  name = "example-tcp-default"
  type = "tcp"
}

resource "bindplane_source" "tcp-custom" {
  rollout = true
  name = "example-tcp-custom"
  type = "tcp"
  parameters_json = jsonencode(
    [
      {
        "name": "listen_ip",
        "value": "0.0.0.0"
      },
      {
        "name": "listen_port",
        "value": 8000
      },
      {
        "name": "log_type",
        "value": "tcp"
      },
      {
        "name": "parse_format",
        "value": "json"
      },
      {
        "name": "regex_pattern",
        "value": ""
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
        "value": "UTC"
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
      },
      {
        "name": "enable_tls",
        "value": true
      },
      {
        "name": "tls_certificate_path",
        "value": "/opt/server/server.crt"
      },
      {
        "name": "tls_private_key_path",
        "value": "/opt/server/server.key"
      },
      {
        "name": "tls_min_version",
        "value": "1.3"
      }
    ]
  )
}
