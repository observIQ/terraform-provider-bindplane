resource "bindplane_fleet" "production" {
  name          = "production-fleet"
  display_name  = "Production Fleet"
  agent_type    = "observiq-otel-collector"
  platform      = "linux"
  configuration = bindplane_configuration.configuration.name
}

resource "bindplane_fleet" "staging" {
  name          = "staging-fleet"
  display_name  = "Staging Fleet"
  agent_type    = "observiq-otel-collector"
  platform      = "linux"
  configuration = "my-config-v2"
}

resource "bindplane_fleet" "development" {
  name          = "development-fleet"
  display_name  = "Development Fleet"
  agent_type    = "observiq-otel-collector"
  platform      = "linux"
  configuration = bindplane_configuration.configuration.name
}
