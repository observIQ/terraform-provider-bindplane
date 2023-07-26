resource "bindplane_source" "custom-default" {
  rollout = true
  name = "example-custom-default"
  type = "host"
}

resource "bindplane_source" "custom-custom" {
  rollout = true
  name = "example-custom-custom"
  type = "host"
  parameters_json = jsonencode(
    [
      {
        "name": "telemetry_types",
        "value": [
          "Traces",
          "Logs",
          "Metrics"
        ]
      },
      {
        "name": "configuration",
        "value": <<EOT
nginx:
  collection_interval: 30s
  endpoint: "http://localhost:80/status"
EOT        
      }
    ]
  )
}
