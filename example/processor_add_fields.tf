resource "bindplane_processor" "add-fields" {
  rollout = false
  name = "example-add-fields"
  type = "add_fields"
  parameters_json = jsonencode(
    [
      {
        "name": "enable_logs",
        "value": true
      },
      {
        "name": "log_condition",
        "value": "true"
      },
      {
        "name": "log_resource_attributes",
        "value": {
          "log-source": "dev",
          "x": "y"
        }
      },
      {
        "name": "log_resource_action",
        "value": "upsert"
      },
      {
        "name": "log_attributes",
        "value": {
          "namespace": "user",
          "user": "dev"
        }
      },
      {
        "name": "log_attributes_action",
        "value": "update"
      },
      {
        "name": "log_body",
        "value": {
          "path": "/v1/api"
        }
      },
      {
        "name": "log_body_action",
        "value": "insert"
      },
      {
        "name": "enable_metrics",
        "value": true
      },
      {
        "name": "datapoint_condition",
        "value": "true"
      },
      {
        "name": "metric_resource_attributes",
        "value": {
          "env": "dev"
        }
      },
      {
        "name": "metric_resource_action",
        "value": "upsert"
      },
      {
        "name": "metric_attributes",
        "value": {
          "allow_error": "false"
        }
      },
      {
        "name": "metric_attributes_action",
        "value": "upsert"
      },
      {
        "name": "enable_traces",
        "value": true
      },
      {
        "name": "span_condition",
        "value": "true"
      },
      {
        "name": "traces_resource_attributes",
        "value": {
          "source": "stage-cluster"
        }
      },
      {
        "name": "traces_resource_action",
        "value": "insert"
      },
      {
        "name": "traces_attributes",
        "value": {}
      },
      {
        "name": "traces_attributes_action",
        "value": "upsert"
      }
    ]
  )
}

