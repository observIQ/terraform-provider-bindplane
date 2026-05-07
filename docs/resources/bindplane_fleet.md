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

| Option          | Type   | Default  | Description                                                                                |
| --------------- | ------ | -------- | ------------------------------------------------------------------------------------------ |
| `name`          | string | required | The resource name for the fleet. Used internally and cannot be changed after creation.     |
| `display_name`  | string | optional | A user-friendly name for the fleet. Can be changed anytime.                               |
| `agent_type`    | string | required | The collector agent type for agents in this fleet. Cannot be changed after creation.       |
| `platform`      | string | required | The platform (OS/architecture) for agents in this fleet. Cannot be changed after creation. |
| `configuration` | string | optional | Name of the configuration assigned to the fleet.                                          |

## Examples

### Basic Fleet with Agent Type and Platform

This example creates a fleet with collector type and platform. The selector is automatically generated to match agents with the fleet name label.

```hcl
resource "bindplane_configuration" "example" {
  name     = "my-config"
  platform = "linux"
  # ... configuration sources, destinations, etc.
}

resource "bindplane_fleet" "production" {
  name          = "us-east"
  display_name  = "US East"
  agent_type    = "observiq-otel-collector"
  platform      = "linux"
  configuration = bindplane_configuration.example.name
}
```

### Fleet with Configuration

This example creates a fleet and assigns a configuration to it.

```hcl
resource "bindplane_fleet" "staging" {
  name          = "staging-fleet"
  display_name  = "Staging Fleet"
  agent_type    = "observiq-otel-collector"
  platform      = "linux"
  configuration = "my-existing-config"
}
```

### Minimal Fleet

This example creates a fleet with minimal required fields (no configuration assigned).

```hcl
resource "bindplane_fleet" "development" {
  name       = "development-fleet"
  agent_type = "observiq-otel-collector"
  platform   = "linux"
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
  agent_type    = "observiq-otel-collector"
  platform      = "linux"
  configuration = "missing-config"  # ← Error: configuration 'missing-config' does not exist
}
```

## Using Fleets

After applying the configuration with `terraform apply`, you can view the fleet with the `bindplane get fleet` commands:

```bash
bindplane get fleet
```

```yaml
# bindplane get fleet us-east -o yaml
apiVersion: bindplane.observiq.com/v1
kind: Fleet
metadata:
    id: us-east
    name: us-east
    displayName: US East
    labels:
        agent-type: observiq-otel-collector
        platform: linux
    version: 1
    dateModified: 2024-05-05T10:30:00Z
spec:
    configuration: my-config
    selector:
        matchLabels:
            fleet: us-east
```

## Import

When using the [terraform import command](https://developer.hashicorp.com/terraform/cli/commands/import),
fleets can be imported by name. For example:

```bash
terraform import bindplane_fleet.production production-fleet
```
