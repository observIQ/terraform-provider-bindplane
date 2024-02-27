resource "bindplane_destination" "dynatrace" {
  rollout = true
  name = "example-dynatrace"
  type = "dynatrace_otlp"
  parameters_json = jsonencode(
    [
      {
        "name": "telemetry_types",
        "value": [
          "Logs",
          "Metrics",
          "Traces"
        ]
      },
      {
        "name": "deployment_type",
        "value": "SaaS"
      },
      {
        "name": "activegate_hostname",
        "value": ""
      },
      {
        "name": "port",
        "value": 9999
      },
      {
        "name": "your_environment_id",
        "value": "abcd"
      },
      {
        "name": "dynatrace_api_token",
        "value": "my-api-token",
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
        "name": "headers",
        "value": {
          "no-cache": "true"
        }
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
