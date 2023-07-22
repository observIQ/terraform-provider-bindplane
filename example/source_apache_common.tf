resource "bindplane_source" "apache-common-default" {
  rollout = true
  name = "example-apache-common-default"
  type = "apache_common"
}

resource "bindplane_source" "apache-common-custom" {
  rollout = true
  name = "example-apache-common-custom"
  type = "apache_common"
  parameters_json = jsonencode(
    [
      {
        "name": "file_path",
        "value": [
          "/var/log/apache2/access.log"
        ]
      },
      {
        "name": "start_at",
        "value": "end"
      }
    ]
  )
}
