resource "bindplane_source" "host-default" {
  rollout = true
  name = "example-host-default"
  type = "host"
}

resource "bindplane_source" "host-custom" {
  rollout = true
  name = "example-host-custom"
  type = "host"
  parameters_json = jsonencode(
    [
      {
        "name": "metric_filtering",
        "value": [
          "system.disk.io",
          "system.disk.io_time",
          "system.disk.merged",
          "system.disk.operation_time",
          "system.disk.operations",
          "system.disk.pending_operations",
          "system.disk.weighted_io_time",
          "system.processes.count",
          "system.processes.created",
          "system.cpu.time",
          "system.cpu.utilization"
        ]
      },
      {
        "name": "enable_process",
        "value": true
      },
      {
        "name": "process_metrics_filtering",
        "value": [
          "process.context_switches"
        ]
      },
      {
        "name": "enable_process_filter",
        "value": true
      },
      {
        "name": "process_include",
        "value": [
          "bindplane-*"
        ]
      },
      {
        "name": "process_exclude",
        "value": [
          "bindplane-agent-*"
        ]
      },
      {
        "name": "process_filter_match_strategy",
        "value": "regexp"
      },
      {
        "name": "collection_interval",
        "value": 30
      }
    ]
  )
}
