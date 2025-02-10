resource "bindplane_source" "otlp" {
  rollout = true
  name    = "example-otlp-default"
  type    = "otlp"
}
