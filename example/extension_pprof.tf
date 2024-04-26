resource "bindplane_extension" "pprof" {
  rollout = true
  name    = "example-pprof"
  type    = "pprof"
  parameters_json = jsonencode(
    [
      {
        "name" : "listen_address",
        "value" : "0.0.0.0"
      },
      {
        "name" : "tcp_port",
        "value" : 5000,
      },
    ]
  )
}
