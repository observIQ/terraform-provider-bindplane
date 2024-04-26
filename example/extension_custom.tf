resource "bindplane_extension" "custom" {
  rollout = true
  name    = "example-custom"
  type    = "custom"
  parameters_json = jsonencode(
    [
      {
        "name" : "telemetry_types",
        "value" : ["Metrics", "Logs", "Traces"]
      },
      {
        "name" : "configuration",
        "value" : "health_check:"
      }
    ]
  )
}
