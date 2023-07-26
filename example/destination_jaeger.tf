resource "bindplane_destination" "jaeger" {
  rollout = true
  name = "example-jaeger"
  type = "jaeger"
  parameters_json = jsonencode(
    [
      {
        "name": "hostname",
        "value": "192.168.10.3"
      },
      {
        "name": "port",
        "value": 14250
      },
      {
        "name": "enable_tls",
        "value": true
      },
      {
        "name": "insecure_skip_verify",
        "value": false
      },
      {
        "name": "ca_file",
        "value": "/opt/tls/ca.crt"
      },
      {
        "name": "mutual_tls",
        "value": true
      },
      {
        "name": "cert_file",
        "value": "/opt/tls/client.crt"
      },
      {
        "name": "key_file",
        "value": "/opt/tls/client.key"
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
      },
      {
        "name": "sending_queue_enabled",
        "value": true
      },
      {
        "name": "sending_queue_num_consumers",
        "value": 10
      },
      {
        "name": "sending_queue_queue_size",
        "value": 5000
      },
      {
        "name": "persistent_queue_enabled",
        "value": true
      },
      {
        "name": "persistent_queue_directory",
        "value": "$OIQ_OTEL_COLLECTOR_HOME/storage"
      }
    ]
  )
}
