resource "bindplane_destination" "splunk-hec" {
  rollout = true
  name = "example-splunk-hec"
  type = "splunkhec"
  parameters_json = jsonencode(
    [
      {
        "name": "token",
        "value": "xxx-xxx-xxx"
      },
      {
        "name": "index",
        "value": "otel"
      },
      {
        "name": "hostname",
        "value": "splunk.corp.net"
      },
      {
        "name": "port",
        "value": 8088
      },
      {
        "name": "path",
        "value": "/services/collector/event"
      },
      {
        "name": "max_request_size",
        "value": 2097152
      },
      {
        "name": "max_event_size",
        "value": 2097152
      },
      {
        "name": "enable_compression",
        "value": true
      },
      {
        "name": "enable_tls",
        "value": true
      },
      {
        "name": "insecure_skip_verify",
        "value": true
      },
      {
        "name": "ca_file",
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
      }
    ]
  )
}
