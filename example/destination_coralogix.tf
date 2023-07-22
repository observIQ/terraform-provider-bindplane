resource "bindplane_destination" "coralogix" {
  rollout = true
  name = "example-coralogix"
  type = "coralogix"
  parameters_json = jsonencode(
    [
      {
        "name": "private_key",
        "value": "xxx-xxx-xxx-xxx",
      },
      {
        "name": "application_name",
        "value": "otel"
      },
      {
        "name": "subsystem_name",
        "value": "otlp"
      },
      {
        "name": "region",
        "value": "EUROPE2"
      },
      {
        "name": "domain",
        "value": ""
      },
      {
        "name": "resource_attributes",
        "value": true
      },
      {
        "name": "application_name_attributes",
        "value": [
          "app",
          "app_name"
        ]
      },
      {
        "name": "subsystem_name_attributes",
        "value": [
          "section"
        ]
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
        "name": "timeout",
        "value": 5
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
