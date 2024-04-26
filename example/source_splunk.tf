resource "bindplane_source" "splunk-default" {
  rollout = true
  name    = "example-splunk-default"
  type    = "splunk_tcp"
}

resource "bindplane_source" "splunk-custom" {
  rollout = true
  name    = "example-splunk-custom"
  type    = "splunk_tcp"
  parameters_json = jsonencode(
    [
      {
        "name" : "listen_ip",
        "value" : "0.0.0.0"
      },
      {
        "name" : "listen_port",
        "value" : 8080
      },
      {
        "name" : "log_type",
        "value" : "splunk_tcp"
      },
      {
        "name" : "parse_format",
        "value" : "regex"
      },
      {
        "name" : "regex_pattern",
        "value" : "^(?P<timestamp>\\d{4}-\\d{2}-\\d{2}\\s+\\d{2}:\\d{2}:\\d{2})\\s+(?P<address>[^\\s]+)\\s+(?P<operation>\\w{3})\\s+(?P<cs_uri_stem>[^\\s]+)\\s(?P<cs_uri_query>[^\\s]+)\\s+(?P<s_port>[^\\s]+)\\s+-\\s+(?P<remoteIp>[^\\s]+)\\s+(?P<userAgent>[^\\s]+)\\s+-\\s+(?P<status>\\d{3})\\s+(?P<sc_status>\\d)\\s+(?P<sc_win32_status>\\d)\\s+(?P<time_taken>[^\\n]+)"
      },
      {
        "name" : "parse_timestamp",
        "value" : true
      },
      {
        "name" : "timestamp_field",
        "value" : "timestamp"
      },
      {
        "name" : "parse_timestamp_format",
        "value" : "ISO8601"
      },
      {
        "name" : "epoch_timestamp_format",
        "value" : "s"
      },
      {
        "name" : "manual_timestamp_format",
        "value" : "%Y-%m-%dT%H:%M:%S.%f%z"
      },
      {
        "name" : "timezone",
        "value" : "UTC"
      },
      {
        "name" : "parse_severity",
        "value" : true
      },
      {
        "name" : "severity_field",
        "value" : "severity"
      },
      {
        "name" : "parse_to",
        "value" : "body"
      },
      {
        "name" : "enable_tls",
        "value" : true
      },
      {
        "name" : "tls_certificate_path",
        "value" : "/opt/tls/server.crt"
      },
      {
        "name" : "tls_private_key_path",
        "value" : "/opt/tls/server.key"
      },
      {
        "name" : "tls_min_version",
        "value" : "1.1"
      }
    ]
  )
}
