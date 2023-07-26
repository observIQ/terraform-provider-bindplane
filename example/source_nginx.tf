resource "bindplane_source" "nginx-default" {
  rollout = true
  name = "example-nginx-default"
  type = "nginx"
}

resource "bindplane_source" "nginx-custom" {
  rollout = true
  name = "example-nginx-custom"
  type = "nginx"
  parameters_json = jsonencode(
    [
      {
        "name": "enable_metrics",
        "value": true
      },
      {
        "name": "endpoint",
        "value": "http://localhost:80/status"
      },
      {
        "name": "disable_metrics",
        "value": []
      },
      {
        "name": "enable_tls",
        "value": true
      },
      {
        "name": "insecure_skip_verify",
        "value": false
      },
      {
        "name": "ca_file",
        "value": "/opt/tls/server-ca.crt"
      },
      {
        "name": "cert_file",
        "value": "/opt/tls/client.crt"
      },
      {
        "name": "key_file",
        "value": "/opt/tls/client.key"
      },
      {
        "name": "collection_interval",
        "value": 30
      },
      {
        "name": "enable_logs",
        "value": true
      },
      {
        "name": "data_flow",
        "value": "high"
      },
      {
        "name": "log_format",
        "value": "observiq"
      },
      {
        "name": "access_log_paths",
        "value": [
          "/var/log/nginx/access.log*"
        ]
      },
      {
        "name": "error_log_paths",
        "value": [
          "/var/log/nginx/error.log*"
        ]
      },
      {
        "name": "start_at",
        "value": "end"
      }
    ]
  )
}
