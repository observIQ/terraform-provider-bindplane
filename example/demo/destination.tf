resource "bindplane_destination" "gateway-east1" {
  rollout = true
  name    = "gateway-east1-tf"
  type    = "bindplane_gateway"
  parameters_json = jsonencode(
    [
      {
        "name" : "hostname",
        "value" : "10.142.0.17"
      },
      {
        "name" : "grpc_port",
        "value" : 4317
      },
      {
        "name" : "protocol",
        "value" : "grpc"
      },
      {
        "name" : "retry_on_failure_enabled",
        "value" : true
      },
      {
        "name" : "retry_on_failure_initial_interval",
        "value" : 1
      },
      {
        "name" : "retry_on_failure_max_interval",
        "value" : 5
      },
      {
        "name" : "retry_on_failure_max_elapsed_time",
        "value" : 10
      },
      {
        "name" : "sending_queue_enabled",
        "value" : true
      },
      {
        "name" : "sending_queue_num_consumers",
        "value" : 2
      },
      {
        "name" : "sending_queue_queue_size",
        "value" : 50
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

resource "bindplane_destination" "nop" {
  rollout = true
  name    = "nop-tf"
  type    = "custom"
  parameters_json = jsonencode(
    [
      {
        "name" : "telemetry_types",
        "value" : ["Metrics", "Logs", "Traces"]
      },
      {
        "name" : "configuration",
        "value" : "nop:"
      }
    ]
  )
}
