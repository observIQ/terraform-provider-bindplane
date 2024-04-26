resource "bindplane_source" "splunkhec-default" {
  rollout = true
  name    = "example-splunkhec-default"
  type    = "splunkhec"
}

resource "bindplane_source" "splunkhec-custom" {
  rollout = true
  name    = "example-splunkhec-custom"
  type    = "splunkhec"
  parameters_json = jsonencode(
    [
      {
        "name" : "listen_ip",
        "value" : "0.0.0.0"
      },
      {
        "name" : "listen_port",
        "value" : 8088
      },
      {
        "name" : "access_token_passthrough",
        "value" : true
      },
      {
        "name" : "enable_tls",
        "value" : true
      },
      {
        "name" : "tls_certificate_path",
        "value" : "/opt/tls/server.crt"
      },
      {
        "name" : "tls_private_key_path",
        "value" : "/opt/tls/server.key"
      }
    ]
  )
}
