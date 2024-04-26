resource "bindplane_destination" "logzio" {
  rollout = true
  name    = "example-logzio"
  type    = "logzio"
  parameters_json = jsonencode(
    [
      {
        "name" : "telemetry_types"
        "value" : [
          "Logs",
          "Metrics",
          "Traces"
        ],
      },
      {
        "name" : "enable_logs",
        "value" : true
      },
      {
        "name" : "logs_token",
        "value" : "xxx-xxx-xxx"
      },
      {
        "name" : "enable_metrics",
        "value" : true
      },
      {
        "name" : "metrics_token",
        "value" : "xxx-xxx-xxx"
      },
      {
        "name" : "listener_url",
        "value" : "https://listener.logz.io:8053"
      },
      {
        "name" : "enable_traces",
        "value" : true
      },
      {
        "name" : "tracing_token",
        "value" : "xxx-xxx-xxx"
      },
      {
        "name" : "region",
        "value" : "us"
      },
      {
        "name" : "timeout",
        "value" : 30
      },
      {
        "name" : "enable_write_ahead_log",
        "value" : true
      },
      {
        "name" : "wal_storage_path",
        "value" : "$OIQ_OTEL_COLLECTOR_HOME/storage/logzio_metrics_wal"
      },
      {
        "name" : "wal_buffer_size",
        "value" : 300
      },
      {
        "name" : "wal_truncate_frequency",
        "value" : 60
      },
      {
        "name" : "retry_on_failure_enabled",
        "value" : true
      },
      {
        "name" : "retry_on_failure_initial_interval",
        "value" : 5
      },
      {
        "name" : "retry_on_failure_max_interval",
        "value" : 30
      },
      {
        "name" : "retry_on_failure_max_elapsed_time",
        "value" : 300
      },
      {
        "name" : "sending_queue_enabled",
        "value" : true
      },
      {
        "name" : "sending_queue_num_consumers",
        "value" : 10
      },
      {
        "name" : "sending_queue_queue_size",
        "value" : 5000
      },
      {
        "name" : "persistent_queue_enabled",
        "value" : true
      },
      {
        "name" : "persistent_queue_directory",
        "value" : "$OIQ_OTEL_COLLECTOR_HOME/storage"
      }
    ]
  )
}
