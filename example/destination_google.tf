resource "bindplane_destination" "google" {
  rollout = true
  name = "example-google"
  type = "googlecloud"
  parameters_json = jsonencode(
    [
      {
        "name": "project",
        "value": "my-project"
      },
      {
        "name": "auth_type",
        "value": "json"
      },
      {
        "name": "credentials",
        "value": <<EOT
{
  "type": "service_account",
  "project_id": "redacted",
  "private_key_id": "redacted",
  "private_key": "redacted",
  "client_email": "redacted",
  "client_id": "redacted",
  "auth_uri": "redacted",
  "token_uri": "redacted",
  "auth_provider_x509_cert_url": "redacted",
  "client_x509_cert_url": "redacted"
}
EOT
      },
      {
        "name": "credentials_file",
        "value": ""
      },
      {
        "name": "retry_on_failure_enabled",
        "value": true
      },
      {
        "name": "retry_on_failure_initial_interval",
        "value": 5
      },
      {
        "name": "retry_on_failure_max_interval",
        "value": 30
      },
      {
        "name": "retry_on_failure_max_elapsed_time",
        "value": 300
      },
      {
        "name": "sending_queue_enabled",
        "value": true
      },
      {
        "name": "sending_queue_num_consumers",
        "value": 10
      },
      {
        "name": "sending_queue_queue_size",
        "value": 5000
      },
      {
        "name": "persistent_queue_enabled",
        "value": true
      },
      {
        "name": "persistent_queue_directory",
        "value": "$OIQ_OTEL_COLLECTOR_HOME/storage"
      },
      {
        "name": "enable_compression",
        "value": true
      },
      {
        "name": "enable_wal",
        "value": true
      },
      {
        "name": "wal_max_backoff",
        "value": 60
      }
    ]
  )
}
