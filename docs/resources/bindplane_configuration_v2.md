---
subcategory: "Pipeline"
description: |-
  A Configuration is a combination of sources, processors, and destinations
  used by BindPlane OP to generate an agent configuration.

  Configuration V2 is the latest iteration of the BindPlane Configuration resource.
  It supports advanced routing and OpenTelemetry Connectors.
---

> [!NOTE]
> bindplane_configuration_v2 resources are supported by BindPlane OP version 1.85.0 and later.

# bindplane_configuration_v2

The `bindplane_configuration_v2` resource creates a BindPlane configuration that can be applied
to one or more managed agents. Configurations are a combination of [sources](./bindplane_source.md),
[destinations](./bindplane_destination.md), [processors](./bindplane_processor.md), and [connectors](./bindplane_connector.md).
Explicit routes can be defined between components to control how telemetry data flows through the configuration.

Configuration V2 builds upon [Configuration V1](./bindplane_configuration.md) by introducing the following features:
- **Routing**: Routes can be defined between all components. This allows for
  more granular control over how telemetry data flows through the configuration.
- **Connectors**: [OpenTelemetry Connectors](https://github.com/open-telemetry/opentelemetry-collector/blob/main/connector/README.md)
  can be defined in the configuration. Connectors are used to route telemetry between pipelines, such as counting the number of logs
  and emitting a metric based on that count.
- **Processors**: Processors can be defined in the configuration without attaching them directly to sources or destinations.
  This allows for more flexibility in how processors are used within a configuration. Processors can still be attached to sources
  or destinations just like before.

## Options

| Option             | Type            | Default  | Description                                                                 |
| ------------------ | --------------- | -------- | --------------------------------------------------------------------------- |
| `name`             | string          | required | The source name.                                                            |
| `platform`         | string          | required | The platform the configuration supports. See the [supported platforms](./bindplane_configuration.md#supported-platforms) section. |
| `labels`           | map             | optional | Key value pairs representing labels to set on the configuration.            |
| `source`           | block           | optional | One or more source blocks. See the [source block](./bindplane_configuration.md#source-block) section. |
| `destination`      | block           | optional | One or more destination blocks. See the [destination block](./bindplane_configuration.md#destination-block) section. |
| `extensions`       | list(string)    | optional | One or more extension names to attach to the configuration.                 |
| `rollout`          | bool            | required | Whether or not updates to the configuration should trigger an automatic rollout of the configuration. |
| `rollout_options`  | block (single)  | optional | Options for configuring the rollout behavior of the configuration. See the [rollout options block](./bindplane_configuration.md#rollout-options-block) section. |

### Source Block

| Option              | Type         | Default  | Description                  |
| ------------------- | ------------ | -------- | ---------------------------- |
| `name`              | string       | required | The source name.             |
| `processors`        | list(string) | optional | One or more processor names to attach to the source. |
| `route`             | string       | optional | One or more routes to attach to the source. See the [route block](./bindplane_configuration.md#route-block) section. |

### Destination Block

| Option              | Type         | Default  | Description                  |
| ------------------- | ------------ | -------- | ---------------------------- |
| `name`              | string       | required | The source name.             |
| `processors`        | list(string) | optional | One or more processor names to attach to the destination. |
| `route_id` | string | required | An arbitrary string that can be used to configure routes to this destination. |

### Rollout Options Block

| Option              | Type         | Default  | Description                  |
| ------------------- | ------------ | -------- | ---------------------------- |
| `type`              | string       | required | The type of rollout to perform. Valid values are 'standard' and 'progressive'. |
| `parameters`        | list(block)  | optional | One or more parameters for the rollout. See the [parameters block](./bindplane_configuration.md#parameters-block) section. |

### Parameters Block

| Option              | Type         | Default  | Description                  |
| ------------------- | ------------ | -------- | ---------------------------- |
| `name`              | string       | required | The name of the parameter.   |
| `value`             | any          | required | The value of the parameter.  |

### Route Block

| Option              | Type         | Default  | Description                  |
| ------------------- | ------------ | -------- | ---------------------------- |
| `telemetry_type`    | enum         | `logs+metrics+traces` | The telemetry type to route. |
| `components`        | list(string) | required | One or more components to route the telemetry to. |

Telemetry types include the following
- `logs`
- `metrics`
- `traces`
- `logs+metrics`
- `logs+traces`
- `metrics+traces`
- `logs+metrics+traces` (default)

The following example shows two route blocks, one for all telemetry types and one for traces only.

```tf
# Route all telemetry types to the Datadog destination
route {
  components = [
    "destinations/${bindplane_destination.datadog.id}"
  ]
}

# Route only traces to the Google destination
route {
  telemetry_type = "traces"
  components = [
    "destinations/${bindplane_destination.google.id}"
  ]
}
```

### Supported Platforms

This table should be used as a reference for supported `platform` values.

| Platform               | Value                   | 
| ---------------------- | ----------------------- |
| Linux                  | `linux`                 |
| Windows                | `windows`               |
| macOS                  | `macos`                 |
| Kubernetes DaemonSet   | `kubernetes-daemonset`  |
| Kubernetes Deployment  | `kubernetes-deployment` |
| OpenShift DaemonSet    | `openshift-daemonset`   |
| OpenShift Deployment   | `openshift-deployment`  |

## Examples

This example shows the creation of a `bindplane_configuration_v2` which uses the following resources:
- two [bindplane_source](./bindplane_source.md) resources
- one [bindplane_destination](./bindplane_destination.md) resource
- two [bindplane_processor](./bindplane_processor.md) resources

Routes are configured on the sources to send telemetry data to the destination.

```hcl
resource "bindplane_source" "host" {
  rollout = true
  name = "my-host"
  type = "host"
  parameters_json = jsonencode(
    [
      {
        "name": "collection_interval",
        "value": 30
      },
      {
        "name": "enable_process",
        "value": false
      }
    ]
  )
}

resource "bindplane_source" "journald" {
  rollout = true
  name = "my-journald"
  type = "journald"
}

resource "bindplane_destination" "google" {
  rollout = true
  name = "my-google"
  type = "googlecloud"
}

resource "bindplane_processor" "batch" {
  rollout = true
  name = "my-batch"
  type = "batch"
}

resource "bindplane_extension" "pprof" {
  rollout = true
  name = "my-pprof"
  type = "pprof"
  parameters_json = jsonencode(
    [
      {
        "name": "listen_address",
        "value": "0.0.0.0"
      },
      {
        "name": "tcp_port",
        "value": 5000,
      }
    ]
  )
}

resource "bindplane_configuration_v2" "configuration" {
  // When removing a component from a configuration and deleting that
  // component during the same apply, we want to update the configuration
  // before the component is deleted.
  lifecycle {
    create_before_destroy = true
  }

  rollout = true
  name = "my-config"
  platform = "linux"
  labels = {
    environment = "production"
    managed-by  = "terraform"
  }

  source {
    name = bindplane_source.host.name
    route {
      components = [
        "destinations/${bindplane_destination.google.id}"
      ]
    }
  }

  source {
    name = bindplane_source.journald.name
    route {
      components = [
        "destinations/${bindplane_destination.google.id}"
      ]
    }
  }

  destination {
    name = bindplane_destination.google.name
    processors = [
      bindplane_processor.batch.name
    ]
  }

  extensions = [
    bindplane_extension.pprof.name
  ]

  rollout_options {
    type = "progressive"
    parameters = [
      {
        name = "stages"
        value = [
          {
            labels = {
              env = "stage"
            }
            name = "stage"
          },
          {
            labels = {
              env = "production"
            }
            name = "production"
          }
        ]
      }
    ]
  }
}
```
