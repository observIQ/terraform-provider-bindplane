resource "bindplane_destination" "custom" {
  rollout = true
  name    = "example-custom"
  type    = "custom"
  parameters_json = jsonencode(
    [
      {
        "name" : "telemetry_types",
        "value" : [
          "Metrics",
          "Logs",
          "Traces"
        ]
      },
      {
        "name" : "configuration",
        "value" : <<EOT
logging:
  verbosity: detailed
  sampling_initial: 5
  sampling_thereafter: 200
EOT
      }
    ]
  )
}
