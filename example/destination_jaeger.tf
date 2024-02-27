resource "bindplane_destination" "jaeger" {
  rollout = true
  name = "example-jaeger"
  type = "jaeger_otlp"
  parameters_json = jsonencode(
    [
      {
        "name": "hostname",
        "value": "jaeger"
      },
      {
        "name": "http_port",
        "value": 4318
      },
      {
        "name": "grpc_port",
        "value": 4317
      },
      {
        "name": "protocol",
        "value": "grpc"
      },
      {
        "name": "http_compression",
        "value": "gzip"
      },
      {
        "name": "grpc_compression",
        "value": "gzip"
      },
      {
        "name": "headers",
        "value": {
          "x-api-key": "xxxx"
        }
      },
      {
        "name": "enable_tls",
        "value": false
      },
      {
        "name": "insecure_skip_verify",
        "value": false
      },
      {
        "name": "ca_file",
        "value": ""
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
