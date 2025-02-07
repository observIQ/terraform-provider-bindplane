resource "bindplane_processor_bundle" "bundle" {
  rollout = true
  name = "my-bundle"

  processor {
    name = bindplane_processor.custom.name
  }

  processor {
    name = bindplane_processor.batch.name
  }
}
