resource "bindplane_source" "fluent-default" {
  rollout = true
  name = "example-fluent-default"
  type = "fluentforward"
}

resource "bindplane_source" "fluent-custom" {
  rollout = true
  name = "example-fluent-custom"
  type = "fluentforward"
  parameters_json = jsonencode(
    [
      {
        "name": "listen_address",
        "value": "0.0.0.0"
      },
      {
        "name": "port",
        "value": 24224
      }
    ]
  )
}
