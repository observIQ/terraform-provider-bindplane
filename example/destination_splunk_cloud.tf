// Splunk Cloud uses the SignalFX exporter.
resource "bindplane_destination" "splunk-cloud" {
  rollout = true
  name = "example-splunk-cloud"
  type = "signalfx"
  parameters_json = jsonencode(
    [
      {
        "name": "token",
        "value": "xxx-xxx-xxx"
      },
      {
        "name": "realm",
        "value": "us0"
      },
      {
        "name": "enable_metrics",
        "value": true
      },
      {
        "name": "enable_logs",
        "value": true
      },
      {
        "name": "enable_traces",
        "value": true
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
