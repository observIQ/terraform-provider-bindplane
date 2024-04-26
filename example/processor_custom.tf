resource "bindplane_processor" "custom" {
  rollout = false
  name    = "example"
  type    = "custom"
  parameters_json = jsonencode(
    [
      {
        "name" : "telemetry_types",
        "value" : [
          "Traces",
          "Logs",
          "Metrics"
        ]
      },
      {
        "name" : "configuration",
        "value" : <<EOT
batch:
  send_batch_size: 100
  send_batch_max_size: 2000
  timeout: 5s\n
EOT
      }
    ]
  )
}

