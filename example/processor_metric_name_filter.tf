resource "bindplane_processor" "metric-name-filter-default" {
  rollout = false
  name = "example-metric-name-filter"
  type = "filter_metric_name"
}

resource "bindplane_processor" "metric-name-filter-custom" {
  rollout = false
  name = "example-metric-name-filter-custom"
  type = "filter_metric_name"
  parameters_json = jsonencode(
    [
      {
        "name": "action",
        "value": "exclude"
      },
      {
        "name": "match_type",
        "value": "regexp"
      },
      {
        "name": "metric_names",
        "value": [
          "system.*",
          "network.*"
        ]
      }
    ]
  )
}

