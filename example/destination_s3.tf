resource "bindplane_destination" "s3" {
  rollout = true
  name = "example-s3"
  type = "aws_s3"
  parameters_json = jsonencode(
    [
      {
        "name": "telemetry_types",
        "value": [
          "Logs",
          "Metrics",
          "Traces"
        ]
      },
      {
        "name": "region",
        "value": "us-east-1"
      },
      {
        "name": "bucket",
        "value": "otel-archive"
      },
      {
        "name": "prefix",
        "value": "otlp"
      },
      {
        "name": "file_prefix",
        "value": "s3-archive"
      },
      {
        "name": "partition",
        "value": "minute"
      }
    ]
  )
}
