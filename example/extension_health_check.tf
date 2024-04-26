resource "bindplane_extension" "health_check" {
  rollout = true
  name    = "example-healthcheck"
  type    = "health_check"
  parameters_json = jsonencode(
    [
      {
        "name" : "listen_address",
        "value" : "0.0.0.0"
      },
      {
        "name" : "listen_port",
        "value" : 8888,
      },
    ]
  )
}
