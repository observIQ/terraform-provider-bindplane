resource "bindplane_destination" "elasticsearch" {
  rollout = true
  name = "example-elasticsearch"
  type = "elasticsearch"
  parameters_json = jsonencode(
    [
      {
        "name": "enable_elastic_cloud",
        "value": false
      },
      {
        "name": "endpoints",
        "value": [
          "https://es-0:9200",
          "https://es-1:9200",
          "https://es-3:9200"
        ]
      },
      {
        "name": "cloudid",
        "value": "my-id"
      },
      {
        "name": "enable_logs",
        "value": true
      },
      {
        "name": "logs_index",
        "value": "otel-logs"
      },
      {
        "name": "enable_traces",
        "value": true
      },
      {
        "name": "traces_index",
        "value": "otel-spans"
      },
      {
        "name": "pipeline",
        "value": "otlp"
      },
      {
        "name": "enable_auth",
        "value": true
      },
      {
        "name": "auth_type",
        "value": "basic"
      },
      {
        "name": "user",
        "value": "otel"
      },
      {
        "name": "password",
        "value": "my-password",
      },
      {
        "name": "api_key",
        "value": "my-api-key",
      },
      {
        "name": "configure_tls",
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
        "name": "num_workers",
        "value": 2
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
