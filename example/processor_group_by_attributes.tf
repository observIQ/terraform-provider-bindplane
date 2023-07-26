resource "bindplane_processor" "group-by-attributes" {
  rollout = false
  name = "example-group-by-attributes"
  type = "group_by_attributes"
  parameters_json = jsonencode(
    [
      {
        "name": "enable_logs",
        "value": true
      },
      {
        "name": "log_attributes",
        "value": [
          "namespace"
        ]
      },
      {
        "name": "enable_metrics",
        "value": true
      },
      {
        "name": "metric_attributes",
        "value": [
          "namespace"
        ]
      },
      {
        "name": "enable_traces",
        "value": true
      },
      {
        "name": "trace_attributes",
        "value": [
          "span_source"
        ]
      }
    ]
  )
}

