resource "bindplane_source" "jvm-default" {
  rollout = true
  name = "example-jvm-default"
  type = "jvm"
}

resource "bindplane_source" "jvm-custom" {
  rollout = true
  name = "example-jvm-custom"
  type = "jvm"
  parameters_json = jsonencode(
    [
      {
        "name": "address",
        "value": "localhost"
      },
      {
        "name": "port",
        "value": 9999
      },
      {
        "name": "jar_path",
        "value": "/opt/opentelemetry-java-contrib-jmx-metrics.jar"
      },
      {
        "name": "collection_interval",
        "value": 60
      }
    ]
  )
}
