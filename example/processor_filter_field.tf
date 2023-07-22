resource "bindplane_processor" "filter-field-default" {
  rollout = false
  name = "example-filter-field"
  type = "filter_field"
}

resource "bindplane_processor" "filter-field-custom" {
  rollout = false
  name = "example-filter-field-custom"
  type = "filter_field"
  parameters_json = jsonencode(
    [
      {
        "name": "telemetry_types",
        "value": [
          "Metrics",
          "Logs",
          "Traces"
        ]
      },
      {
        "name": "action",
        "value": "exclude"
      },
      {
        "name": "match_type",
        "value": "strict"
      },
      {
        "name": "resources",
        "value": {
          "k8s.namespace.name": "dev"
        }
      },
      {
        "name": "attributes",
        "value": {
          "env": "stage"
        }
      },
      {
        "name": "bodies",
        "value": {
          "path": "/health"
        }
      }
    ]
  )
}

