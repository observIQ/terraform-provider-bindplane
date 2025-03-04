resource "bindplane_source" "file-default" {
  rollout = true
  name    = "example-file-default"
  type    = "file_v2"
}

resource "bindplane_source" "file-custom" {
  rollout = true
  name    = "example-file-custom"
  type    = "file_v2"
  parameters_json = jsonencode(
    [
      {
        "name" : "file_path",
        "value" : [
          "/var/log/bindplane/bindplane.*"
        ],
      },
      {
        "name" : "exclude_file_path",
        "value" : [
          "*.gz"
        ],
      },
      {
        "name" : "log_type",
        "value" : "terraform-file",
      },
      {
        "name" : "parse_format",
        "value" : "regex",
      },
      {
        "name" : "regex_pattern",
        "value" : "^(?P<timestamp>\\d{4}-\\d{2}-\\d{2}\\s+\\d{2}:\\d{2}:\\d{2})\\s+(?P<address>[^\\s]+)\\s+(?P<operation>\\w{3})\\s+(?P<cs_uri_stem>[^\\s]+)\\s(?P<cs_uri_query>[^\\s]+)\\s+(?P<s_port>[^\\s]+)\\s+-\\s+(?P<remoteIp>[^\\s]+)\\s+(?P<userAgent>[^\\s]+)\\s+-\\s+(?P<status>\\d{3})\\s+(?P<sc_status>\\d)\\s+(?P<sc_win32_status>\\d)\\s+(?P<time_taken>[^\n]+)",
      },
      {
        "name" : "multiline_parsing",
        "value" : "specify line end",
      },
      {
        "name" : "multiline_line_start_pattern",
        "value" : "",
      },
      {
        "name" : "multiline_line_end_pattern",
        "value" : "END",
      },
      {
        "name" : "parse_timestamp",
        "value" : true,
      },
      {
        "name" : "timestamp_field",
        "value" : "timestamp",
      },
      {
        "name" : "parse_timestamp_format",
        "value" : "Manual",
      },
      {
        "name" : "epoch_timestamp_format",
        "value" : "s",
      },
      {
        "name" : "manual_timestamp_format",
        "value" : "%Y-%m-%d %H:%M:%S",
      },
      {
        "name" : "timezone",
        "value" : "America/Detroit",
      },
      {
        "name" : "parse_severity",
        "value" : true,
      },
      {
        "name" : "severity_field",
        "value" : "status",
      },
      {
        "name" : "include_file_name_attribute",
        "value" : true,
      },
      {
        "name" : "include_file_path_attribute",
        "value" : false,
      },
      {
        "name" : "include_file_name_resolved_attribute",
        "value" : true,
      },
      {
        "name" : "include_file_path_resolved_attribute",
        "value" : false,
      },
      {
        "name" : "encoding",
        "value" : "ascii",
      },
      {
        "name" : "poll_interval",
        "value" : 500,
      },
      {
        "name" : "max_concurrent_files",
        "value" : 3,
      },
      {
        "name" : "parse_to",
        "value" : "attributes",
      },
      {
        "name" : "start_at",
        "value" : "end",
      },
      {
        "name" : "fingerprint_size",
        "value" : "2kb",
      },
      {
        "name" : "enable_offset_storage",
        "value" : false,
      },
      {
        "name" : "offset_storage_dir",
        "value" : "/opt/storage",
      },
      {
        "name" : "retry_on_failure_enabled",
        "value" : false,
      },
      {
        "name" : "retry_on_failure_initial_interval",
        "value" : 2,
      },
      {
        "name" : "retry_on_failure_max_interval",
        "value" : 10,
      },
      {
        "name" : "retry_on_failure_max_elapsed_time",
        "value" : 60,
      },
      {
        "name" : "enable_sorting",
        "value" : false,
      },
      {
        "name" : "sorting_regex",
        "value" : "^*",
      },
      {
        "name" : "sort_rules",
        "value" : [],
      }
    ]
  )
}
