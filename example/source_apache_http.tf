resource "bindplane_source" "apache-default" {
  rollout = true
  name    = "example-apache-default"
  type    = "apache_http"
}

resource "bindplane_source" "apache-custom" {
  rollout = true
  name    = "example-apache-custom"
  type    = "apache_http"
  parameters_json = jsonencode(
    [
      {
        "name" : "telemetry_types"
        "value" : [
          "Logs",
          "Metrics",
        ],
      },
      {
        "name" : "enable_metrics",
        "value" : true
      },
      {
        "name" : "hostname",
        "value" : "localhost"
      },
      {
        "name" : "port",
        "value" : 80
      },
      {
        "name" : "collection_interval",
        "value" : 60
      },
      {
        "name" : "enable_tls",
        "value" : false
      },
      {
        "name" : "strict_tls_verify",
        "value" : false
      },
      {
        "name" : "ca_file",
        "value" : ""
      },
      {
        "name" : "mutual_tls",
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
        "name" : "disable_metrics",
        "value" : []
      },
      {
        "name" : "enable_logs",
        "value" : true
      },
      {
        "name" : "access_log_path",
        "value" : [
          "/var/log/apache2/access.log"
        ]
      },
      {
        "name" : "error_log_path",
        "value" : [
          "/var/log/apache2/error.log"
        ]
      },
      {
        "name" : "timezone",
        "value" : "UTC"
      },
      {
        "name" : "start_at",
        "value" : "end"
      }
    ]
  )
}
