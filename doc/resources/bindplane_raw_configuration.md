# bindplane_raw_configuration

The `bindplane_raw_configuration` resource takes a raw OpenTelemetry configuration and applys it
to BindPlane. If the `match_labels` match an agent, BindPlane will push the configuration to the agent(s).

| Option              | Type   | Default  | Description                  |
| ------------------- | -----  | -------- | ---------------------------- |
| `name`              | string | required | The configuration name.                                       |
| `platform`     | string  | required | The platform the configuration supports. See the [supported platforms](./bindplane_configuration.md#supported-platforms) section. |
| `labels`       | map     | optional | Key value pairs representing labels to set on the configuration. |
| `raw_configuration` | string | required | The OpenTelemetry configuration that will be applied to agents |
| `rollout`      | bool    | required | Whether or not updates to the configuration should trigger an automatic rollout of the configuration. |

The following example will match agents with labels `env=stage,platform=frontend`.

```tf
resource "bindplane_raw_configuration" "raw" {
  rollout = true
  name = "raw"
  platform = "linux"
  labels = {
    env = "stage",
    platform = "frontend"
  }
  raw_configuration = <<EOT
receivers:
  prometheus:
    config:
      scrape_configs:
        - job_name: 'collector'
          scrape_interval: 10s
          static_configs:
            - targets:
                - 'localhost:8888'
exporters:
  logging:
service:
  pipelines:
    metrics:
      receivers:
        - prometheus
      exporters:
        - logging
EOT
}
```
