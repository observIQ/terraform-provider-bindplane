resource "bindplane_destination" "zipkin" {
  rollout = true
  name    = "example-zipkin"
  type    = "zipkin"
  parameters_json = jsonencode(
    [
      {
        "name" : "hostname",
        "value" : "spans.corp.net"
      },
      {
        "name" : "port",
        "value" : 9411
      },
      {
        "name" : "path",
        "value" : "/api/v2/spans"
      },
      {
        "name" : "enable_tls",
        "value" : true
      },
      {
        "name" : "insecure_skip_verify",
        "value" : true
      },
      {
        "name" : "ca_file",
        "value" : ""
      },
      {
        "name" : "mutual_tls",
        "value" : false
      },
      {
        "name" : "cert_file",
        "value" : ""
      },
      {
        "name" : "key_file",
        "value" : ""
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
