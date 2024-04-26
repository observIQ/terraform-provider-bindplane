resource "bindplane_source" "syslog-default" {
  rollout = true
  name    = "example-syslog-default"
  type    = "syslog"
}

resource "bindplane_source" "syslog-custom" {
  rollout = true
  name    = "example-syslog-custom"
  type    = "syslog"
  parameters_json = jsonencode(
    [
      {
        "name" : "listen_ip",
        "value" : "0.0.0.0"
      },
      {
        "name" : "listen_port",
        "value" : 5140
      },
      {
        "name" : "protocol",
        "value" : "rfc5424"
      },
      {
        "name" : "connection_type",
        "value" : "tcp"
      },
      {
        "name" : "data_flow",
        "value" : "high"
      },
      {
        "name" : "timezone",
        "value" : "America/Detroit"
      },
      {
        "name" : "parse_to",
        "value" : "body"
      },
      {
        "name" : "enable_octet_counting",
        "value" : false
      },
      {
        "name" : "enable_non_transparent_framing_trailer",
        "value" : true
      },
      {
        "name" : "non_transparent_framing_trailer",
        "value" : "NUL"
      },
      {
        "name" : "enable_mutual_tls",
        "value" : false
      },
      {
        "name" : "cert_file",
        "value" : ""
      },
      {
        "name" : "key_file",
        "value" : ""
      },
      {
        "name" : "ca_file",
        "value" : ""
      },
      {
        "name" : "tls_min_version",
        "value" : "1.2"
      }
    ]
  )
}
