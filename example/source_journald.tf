resource "bindplane_source" "journald-default" {
  rollout = true
  name = "example-journald-default"
  type = "journald"
}

resource "bindplane_source" "journald-custom" {
  rollout = true
  name = "example-journald-custom"
  type = "journald"
  parameters_json = jsonencode(
    [
      {
        "name": "units",
        "value": [
          "bindplane",
          "observiq-otel-collector",
          "nginx"
        ]
      },
      {
        "name": "directory",
        "value": "/run/log/journa"
      },
      {
        "name": "priority",
        "value": "warning"
      },
      {
        "name": "start_at",
        "value": "beginning"
      }
    ]
  )
}
