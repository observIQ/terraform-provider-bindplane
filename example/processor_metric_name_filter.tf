resource "bindplane_processor" "metric-name-filter" {
  rollout = false
  name    = "example-metric-name-filter"
  type    = "filter_metric_name"
  parameters_json = jsonencode(
    [
      {
        "name" : "action",
        "value" : "exclude"
      },
      {
        "name" : "match_type",
        "value" : "regexp"
      },
      {
        "name" : "metric_names",
        "value" : [
          "system.*",
          "network.*"
        ]
      }
    ]
  )
}

