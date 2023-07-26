resource "bindplane_destination" "otlp" {
  rollout = true
  name = "example-otlp"
  type = "otlp_grpc"
  parameters_json = jsonencode(
    [
      {
        "name": "hostname",
        "value": "otlp.corp.net"
      },
      {
        "name": "grpc_port",
        "value": 4317
      },
      {
        "name": "http_port",
        "value": 4318
      },
      {
        "name": "protocol",
        "value": "grpc"
      },
      {
        "name": "headers",
        "value": {
          "env": "dev",
          "token": "xxx-xxx-xxx"
        }
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
