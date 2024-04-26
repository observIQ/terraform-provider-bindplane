resource "bindplane_processor" "rename-fields" {
  rollout = false
  name    = "example-rename-fields"
  type    = "rename_field"
  parameters_json = jsonencode(
    [
      {
        "name" : "telemetry_types"
        "value" : [
          "Logs",
          "Metrics",
          "Traces"
        ],
      },
      {
        "name" : "enable_logs",
        "value" : true
      },
      {
        "name" : "log_condition",
        "value" : "true"
      },
      {
        "name" : "log_resource_keys",
        "value" : {
          "namespace" : "k8s.namespace.name"
        }
      },
      {
        "name" : "log_attribute_keys",
        "value" : {
          "auth" : "user"
        }
      },
      {
        "name" : "log_body_keys",
        "value" : {
          "api_path" : "path"
        }
      },
      {
        "name" : "enable_metrics",
        "value" : true
      },
      {
        "name" : "datapoint_condition",
        "value" : "true"
      },
      {
        "name" : "metric_resource_keys",
        "value" : {
          "host.name" : "host"
        }
      },
      {
        "name" : "metric_attribute_keys",
        "value" : {}
      },
      {
        "name" : "enable_traces",
        "value" : true
      },
      {
        "name" : "span_condition",
        "value" : "true"
      },
      {
        "name" : "trace_resource_keys",
        "value" : {}
      },
      {
        "name" : "trace_attribute_keys",
        "value" : {
          "id" : "span_id"
        }
      }
    ]
  )
}

