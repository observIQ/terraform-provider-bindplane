resource "bindplane_fleet" "production" {
  name        = "production-fleet"
  description = "Fleet for production agents"
  configuration = bindplane_configuration.configuration.name

  labels = {
    environment = "production"
    team        = "platform"
  }

  selector {
    match_labels = {
      fleet = "production-fleet"
      env   = "prod"
    }
  }
}

resource "bindplane_fleet" "staging" {
  name          = "staging-fleet-2"
  description   = "Fleet for staging agents"
  configuration = bindplane_configuration_v2.configuration.name

  labels = {
    environment = "staging"
    team        = "platform"
  }

  selector {
    match_labels = {
      fleet = "staging-fleet"
      env   = "staging"
    }
  }
}

resource "bindplane_fleet" "development" {
  name        = "development-fleet"
  description = "Fleet for development agents"
  configuration = bindplane_configuration.configuration.name

  labels = {
    environment = "development"
    team        = "platform"
  }
}
