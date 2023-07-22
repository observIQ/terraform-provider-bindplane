resource "bindplane_destination" "grafana" {
  rollout = true
  name = "example-grafana"
  type = "grafana_cloud"
  parameters_json = jsonencode(
    [
      {
        "name": "metric_instance_id",
        "value": "id-xx-xxx-xx"
      },
      {
        "name": "api_key",
        "value": "xxx-xxx-xxx"
      },
      {
        "name": "gcp_zone",
        "value": "prod-us-central-0"
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
