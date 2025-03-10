resource "bindplane_configuration" "configuration-full" {
  // When removing a component from a configuration and deleting that
  // component during the same apply, we want to update the configuration
  // before the component is deleted.
  lifecycle {
    create_before_destroy = true
  }

  // Automatically rollout new versions of the configuration
  // to managed agents. This includes changes to the underlying
  // sources, processors, and destinations.
  rollout = true

  name = "example-configuration-full"

  // Linux supports most sources. Other options include
  // macos and windows.
  platform = "linux"

  // Optional labels
  labels = {
    managed = "terraform"
    //repo = "https://github.com/observIQ/terraform-provider-bindplane"
    purpose = "example"
  }

  // One or more source blocks can be configured here.
  // Sources are configured with optional processors.

  source {
    name = bindplane_source.otlp-custom.name
  }

  source {
    name = bindplane_source.host-custom.name
    processors = [
      // Use filter processor to omit metrics by name.
      bindplane_processor.metric-name-filter.name,
    ]
  }

  source {
    name = bindplane_source.file-custom.name
  }

  // One or more destination blocks can be configured here.
  // Destinations are configured with optional processors.

  // Send metrics, traces, and logs to Grafana Cloud.
  destination {
    name = bindplane_destination.grafana.name
    processors = [
      // Batch and group telemetry before sending to Grafana Cloud.
      bindplane_processor.batch.name,
      bindplane_processor.group-by-attributes.name
    ]
  }

  // The custom destination implements the logging exporter,
  // for logging telemetry to the collectors log file.
  destination {
    name = bindplane_destination.custom.name
  }

  extensions = [
    bindplane_extension.health_check.name,
    bindplane_extension.pprof.name,
    bindplane_extension.custom.name
  ]
}
