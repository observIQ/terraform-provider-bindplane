resource "bindplane_destination" "prometheus" {
  rollout = true
  name = "example-prometheus"
  type = "prometheus"
  parameters_json = jsonencode(
    [
      {
        "name": "listen_address",
        "value": "0.0.0.0"
      },
      {
        "name": "listen_port",
        "value": 9000
      },
      {
        "name": "namespace",
        "value": "dev"
      },
      {
        "name": "configure_tls",
        "value": true
      },
      {
        "name": "cert_file",
        "value": "/opt/tls/server.crt"
      },
      {
        "name": "key_file",
        "value": "/opt/tls/server.key"
      },
      {
        "name": "ca_file",
        "value": ""
      },
      {
        "name": "mutual_tls",
        "value": false
      },
      {
        "name": "client_ca_file",
        "value": ""
      }
    ]
  )
}
