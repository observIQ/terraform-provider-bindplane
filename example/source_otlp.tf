resource "bindplane_source" "otlp-default" {
  rollout = true
  name    = "example-otlp-default"
  type    = "otlp"
}

resource "bindplane_source" "otlp-custom" {
  rollout = true
  name    = "example-otlp-custom"
  type    = "otlp"
  parameters_json = jsonencode(
    [
      {
        "name" : "listen_address",
        "value" : "0.0.0.0"
      },
      {
        "name" : "http_port",
        "value" : 4317
      },
      {
        "name" : "grpc_port",
        "value" : 4318
      },
      {
        "name" : "enable_tls",
        "value" : true
      },
      {
        "name" : "mutual_tls",
        "value" : true
      },
      {
        "name" : "ca_file",
        "value" : "/etc/otel/bindplane-ca.crt"
      },
      {
        "name" : "cert_file",
        "value" : "/etc/otel/bindplane-client.crt"
      },
      {
        "name" : "key_file",
        "value" : "/etc/otel/bindplane-client.key"
      },
      {
        "name" : "enable_grpc_timeout",
        "value" : true
      },
      {
        "name" : "grpc_max_connection_idle",
        "value" : 20,
      },
      {
        "name" : "grpc_max_connection_age",
        "value" : 60
      },
      {
        "name" : "grpc_max_connection_age_grace",
        "value" : 120
      },
    ]
  )
}
