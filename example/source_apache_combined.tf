resource "bindplane_source" "apache-combined-default" {
  rollout = true
  name = "example-apache-combined-default"
  type = "apache_combined"
}

resource "bindplane_source" "apache-combined-custom" {
  rollout = true
  name = "example-apache-combined-custom"
  type = "apache_combined"
  parameters_json = jsonencode(
    [
      {
        "name": "file_path",
        "value": [
          "/var/log/apache_combined.log"
        ]
      },
      {
        "name": "parse_to",
        "value": "body"
      },
      {
        "name": "start_at",
        "value": "end"
      }
    ]
  )
}
