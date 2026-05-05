---
subcategory: "Fleet Management"
description: |-
  A Fleet manages a collection of agents that are deployed and configured by Bindplane.
---

# bindplane_fleet

The `bindplane_fleet` resource creates a Bindplane fleet that manages a collection of agents.
A fleet defines which [configuration](./bindplane_configuration.md) is assigned to a group of agents,
and uses a selector to match agents to the fleet based on labels.

## Options

| Option          | Type   | Default  | Description                                                       |
| --------------- | ------ | -------- | ----------------------------------------------------------------- |
| `name`          | string | required | The fleet name.                                                   |
| `description`   | string | optional | The fleet description.                                            |
| `configuration` | string | optional | Name of the configuration assigned to the fleet.                  |
| `labels`        | map    | optional | Labels to assign to the fleet for organization and filtering.     |
| `selector`      | object | optional | Agent selector for matching agents to this fleet.                 |

### Selector

The `selector` block matches agents to the fleet based on labels.

| Option         | Type   | Default  | Description                                      |
| -------------- | ------ | -------- | ------------------------------------------------ |
| `match_labels` | map    | optional | Map of label key-value pairs to match agents.    |

## Examples

### Basic Fleet

This example creates a fleet that assigns a configuration to agents with matching labels.

```hcl
resource "bindplane_configuration" "example" {
  name     = "my-config"
  platform = "linux"
  # ... configuration sources, destinations, etc.
}

resource "bindplane_fleet" "production" {
  name          = "production-fleet"
  description   = "Fleet for production agents"
  configuration = bindplane_configuration.example.name

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
```

### Fleet with v2 Configuration

This example uses a v2 configuration with the fleet.

```hcl
resource "bindplane_fleet" "staging" {
  name          = "staging-fleet"
  description   = "Fleet for staging agents"
  configuration = bindplane_configuration_v2.example.name

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
```

### Fleet without Selector

This example creates a fleet without a selector, which will not match any agents.

```hcl
resource "bindplane_fleet" "development" {
  name          = "development-fleet"
  description   = "Fleet for development agents"
  configuration = bindplane_configuration.example.name

  labels = {
    environment = "development"
  }
}
```

## Configuration Dependency

When you reference a configuration from another resource using its name attribute (e.g., `bindplane_configuration.example.name`),
Terraform automatically establishes a dependency. The configuration will be created before the fleet, ensuring the referenced
configuration exists when the fleet is created.

If you hardcode a configuration name as a string literal, the provider validates that the configuration exists at creation time
and returns an error if it doesn't:

```hcl
# This will fail if "missing-config" doesn't exist
resource "bindplane_fleet" "example" {
  name          = "my-fleet"
  configuration = "missing-config"  # ← Error: configuration 'missing-config' does not exist
}
```

## Using Fleets

After applying the configuration with `terraform apply`, you can view the fleet with the `bindplane get fleet` commands:

```bash
bindplane get fleet
```

```yaml
# bindplane get fleet production-fleet -o yaml
apiVersion: bindplane.observiq.com/v1
kind: Fleet
metadata:
    id: 01HQNNG5JFCY74WQ4MAEVD61H7
    name: production-fleet
    description: Fleet for production agents
    labels:
        environment: production
        team: platform
    version: 1
    dateModified: 2024-05-05T10:30:00Z
spec:
    configuration: my-config
    selector:
        matchLabels:
            fleet: production-fleet
            env: prod
```

## Import

When using the [terraform import command](https://developer.hashicorp.com/terraform/cli/commands/import),
fleets can be imported by name. For example:

```bash
terraform import bindplane_fleet.production production-fleet
```
