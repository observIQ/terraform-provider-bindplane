resource "bindplane_destination" "kafka" {
  rollout = true
  name = "example-kafka"
  type = "kafka_otlp_destination"
  parameters_json = jsonencode(
    [
      {
        "name": "protocol_version",
        "value": "2.2.1"
      },
      {
        "name": "brokers",
        "value": [
          "kafka-0:9092",
          "kafka-1:9092"
        ]
      },
      {
        "name": "timeout",
        "value": 5
      },
      {
        "name": "enable_auth",
        "value": true
      },
      {
        "name": "auth_type",
        "value": "tls"
      },
      {
        "name": "basic_username",
        "value": ""
      },
      {
        "name": "basic_password",
        "value": "(sensitive)",
        "sensitive": true
      },
      {
        "name": "sasl_username",
        "value": ""
      },
      {
        "name": "sasl_password",
        "value": "(sensitive)",
        "sensitive": true
      },
      {
        "name": "sasl_mechanism",
        "value": "SCRAM-SHA-256"
      },
      {
        "name": "tls_insecure",
        "value": false
      },
      {
        "name": "tls_ca_file",
        "value": "/opt/tls/ca.crt"
      },
      {
        "name": "tls_cert_file",
        "value": "/opt/tls/client.crt"
      },
      {
        "name": "tls_key_file",
        "value": "/opt/tls/client.key"
      },
      {
        "name": "tls_server_name_override",
        "value": ""
      },
      {
        "name": "kerberos_service_name",
        "value": ""
      },
      {
        "name": "kerberos_realm",
        "value": ""
      },
      {
        "name": "kerberos_config_file",
        "value": "/etc/krb5.conf"
      },
      {
        "name": "kerberos_auth_type",
        "value": "keytab"
      },
      {
        "name": "kerberos_keytab_file",
        "value": "/etc/security/kafka.keytab"
      },
      {
        "name": "kerberos_username",
        "value": ""
      },
      {
        "name": "kerberos_password",
        "value": "(sensitive)",
        "sensitive": true
      },
      {
        "name": "enable_metrics",
        "value": true
      },
      {
        "name": "metric_topic",
        "value": "otlp_metrics"
      },
      {
        "name": "enable_logs",
        "value": true
      },
      {
        "name": "log_topic",
        "value": "otlp_logs"
      },
      {
        "name": "enable_traces",
        "value": true
      },
      {
        "name": "trace_topic",
        "value": "otlp_spans"
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
      }
    ]
  )
}
