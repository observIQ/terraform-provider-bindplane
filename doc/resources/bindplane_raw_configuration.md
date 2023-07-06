# bindplane_raw_configuration

The `bindplane_raw_configuration` resource takes a raw OpenTelemetry configuration and applys it
to BindPlane. If the `match_labels` match an agent, BindPlane will push the configuration to the agent(s).

| Option              | Type   | Default  | Description                  |
| ------------------- | -----  | -------- | ---------------------------- |
| `name`              | string | required | The configuration name                                         |
| `labels`            | map    | required | Friendly labels                                                |
| `match_labels`      | map    | required | The labels that will be used for matching agents               |
| `raw_configuration` | string | required | The OpenTelemetry configuration that will be applied to agents |

The following example will match agents with labels `env=stage,platform=frontend`.

```tf
resource "bindplane_raw_configuration" "config" {
  name = "stage"
  labels = {
    env = "stage"
  }
  match_labels = {
    env = "stage"
    platform = "frontend"
  }
  raw_configuration = <<EOT
receivers:
  hostmetrics:
    collection_interval: 60s
    scrapers:
      cpu:
exporters:
  logging:
service:
  pipelines:
    metrics:
      receivers: [hostmetrics]
      exporters: [logging]
EOT
}
```
