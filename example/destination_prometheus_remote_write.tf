resource "bindplane_destination" "prometheus-remote-write" {
  rollout = true
  name = "example-prometheus-remote-write"
  type = "prometheus_remote_write"
  parameters_json = jsonencode(
    [
      {
        "name": "hostname",
        "value": "https://mezmo.corp.net"
      },
      {
        "name": "port",
        "value": 9009
      },
      {
        "name": "path",
        "value": "/v1/metrics"
      },
      {
        "name": "namespace",
        "value": "dev"
      },
      {
        "name": "enable_resource_to_telemetry_conversion",
        "value": true
      },
      {
        "name": "headers",
        "value": {
          "token": "xxx-xxx-xxx"
        }
      },
      {
        "name": "external_labels",
        "value": {
          "namespace": "otel"
        }
      },
      {
        "name": "enable_tls",
        "value": true
      },
      {
        "name": "strict_tls_verify",
        "value": true
      },
      {
        "name": "ca_file",
        "value": "/opt/tls/ca.crt"
      },
      {
        "name": "mutual_tls",
        "value": false
      },
      {
        "name": "cert_file",
        "value": ""
      },
      {
        "name": "key_file",
        "value": ""
      },
      {
        "name": "enable_write_ahead_log",
        "value": true
      },
      {
        "name": "wal_storage_path",
        "value": "prometheus_rw"
      },
      {
        "name": "wal_buffer_size",
        "value": 300
      },
      {
        "name": "wal_truncate_frequency",
        "value": 60
      },
      {
        "name": "retry_on_failure_enabled",
        "value": true
      },
      {
        "name": "retry_on_failure_initial_interval",
        "value": 5
      },
      {
        "name": "retry_on_failure_max_interval",
        "value": 30
      },
      {
        "name": "retry_on_failure_max_elapsed_time",
        "value": 300
      }
    ]
  )
}
