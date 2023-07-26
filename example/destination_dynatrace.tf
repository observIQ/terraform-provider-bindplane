resource "bindplane_destination" "dynatrace" {
  rollout = true
  name = "example-dynatrace"
  type = "dynatrace"
  parameters_json = jsonencode(
    [
      {
        "name": "metric_ingest_endpoint",
        "value": "https://my-corp/e/dev/api/v2/metrics/ingest"
      },
      {
        "name": "api_token",
        "value": "xxx-xxx-xxx",
      },
      {
        "name": "resource_to_telemetry_conversion",
        "value": true
      },
      {
        "name": "prefix",
        "value": "otel"
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
