---
subcategory: "Pipeline"
description: |-
  A Configuration is a combination of sources, processors, and destinations
  used by Bindplane to generate an agent configuration.

  Configuration V2 is the latest iteration of the Bindplane Configuration resource.
  It supports advanced routing and OpenTelemetry Connectors.
---

> [!NOTE]
> bindplane_configuration_v2 resources are supported by Bindplane version 1.85.0 and later.

# bindplane_configuration_v2

The `bindplane_configuration_v2` resource creates a Bindplane configuration that can be applied
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
| `description`      | string          | optional | The description of the configuration, displayed in the Bindplane UI. Sets the configuration's `metadata.description` field. |
| `labels`           | map             | optional | Key value pairs representing labels to set on the configuration. Set `role = "gateway"` to mark a gateway configuration. See [gateway configurations](#gateway-configurations). |
| `source`           | block           | optional | One or more source blocks. See the [source block](#source-block) section. |
| `connector`        | block           | optional | One or more connector blocks. See the [connector block](#connector-block) section. |
| `processor_group`  | block           | optional | One or more processor group blocks. See the [processor group block](#processor-group-block) section. |
| `destination`      | block           | optional | One or more destination blocks. See the [destination block](#destination-block) section. |
| `extensions`       | list(string)    | optional | One or more extension names to attach to the configuration.                 |
| `rollout`          | bool            | required | Whether or not updates to the configuration should trigger an automatic rollout of the configuration. |
| `rollout_options`  | block (single)  | optional | Options for configuring the rollout behavior of the configuration. See the [rollout options block](./bindplane_configuration.md#rollout-options-block) section. |
| `advanced`         | block (single)  | optional | Advanced configuration options. See the [advanced section](#advanced) below. |

### Source Block

| Option              | Type         | Default  | Description                  |
| ------------------- | ------------ | -------- | ---------------------------- |
| `name`              | string       | required | The source name.             |
| `processors`        | list(string) | optional | One or more processor names to attach to the source. Conflicts with `processor`. |
| `processor`         | block        | optional | One or more processors to attach to the source, referenced by name or defined inline. Conflicts with `processors`. See the [processor block](#processor-block) section. |
| `route`             | string       | optional | One or more routes to attach to the source. See the [route block](#route-block) section. |

### Connector Block

A connector is either a [bindplane_connector](./bindplane_connector.md) referenced by `name`, or
an inline connector defined by `type` and `parameters_json`.

| Option              | Type         | Default  | Description                  |
| ------------------- | ------------ | -------- | ---------------------------- |
| `route_id`          | string       | required | An arbitrary string that can be used to configure routes to this connector. |
| `name`              | string       | optional | The connector name. Conflicts with `type`. |
| `type`              | string       | optional | The type of an inline connector, such as `routing`. Conflicts with `name`. |
| `parameters_json`   | string       | optional | A JSON array of parameters used to configure an inline connector. Requires `type`. |
| `route`             | string       | optional | One or more routes to attach to the connector. See the [route block](#route-block) section. |

When a connector routes telemetry with a `routes` parameter, such as the `routing` connector, Bindplane
matches each `route` block to the entry in the `routes` parameter with the same ID. The `route_id` of each
`route` block must be the `id` of an entry in the `routes` parameter.

### Processor Group Block

| Option              | Type         | Default  | Description                  |
| ------------------- | ------------ | -------- | ---------------------------- |
| `route_id`          | string       | required | An arbitrary string that can be used to configure routes to this processor group. |
| `parameters_json`   | string       | optional | A JSON array of parameters used to configure the processor group, such as `telemetry_types`. |
| `processors`        | list(string) | optional | One or more processor names to attach to the processor group. Conflicts with `processor`. |
| `processor`         | block        | optional | One or more processors to attach to the processor group, referenced by name or defined inline. Conflicts with `processors`. See the [processor block](#processor-block) section. |
| `route`             | string       | optional | One or more routes to attach to the processor group. See the [route block](#route-block) section. |

### Destination Block

| Option              | Type         | Default  | Description                  |
| ------------------- | ------------ | -------- | ---------------------------- |
| `route_id`          | string       | required | An arbitrary string that can be used to configure routes to this destination. |
| `name`              | string       | required | The source name.             |
| `processors`        | list(string) | optional | One or more processor names to attach to the destination. Conflicts with `processor`. |
| `processor`         | block        | optional | One or more processors to attach to the destination, referenced by name or defined inline. Conflicts with `processors`. See the [processor block](#processor-block) section. |

### Processor Block

A processor is either a [bindplane_processor](./bindplane_processor.md) referenced by `name`, or
an inline processor defined by `type` and `parameters_json`. Processors run in the order they are defined.

| Option              | Type         | Default  | Description                  |
| ------------------- | ------------ | -------- | ---------------------------- |
| `name`              | string       | optional | The processor name. Conflicts with `type`. |
| `type`              | string       | optional | The type of an inline processor, such as `batch`. Conflicts with `name`. |
| `parameters_json`   | string       | optional | A JSON array of parameters used to configure an inline processor. Requires `type`. |
| `display_name`      | string       | optional | The name of the processor displayed in the Bindplane UI. |
| `id`                | string       | optional | The ID of an inline processor. Generated by Bindplane when not set. |

Bindplane generates an ID for each inline processor that does not have one. The provider saves the ID
to state and sends it on later applies, so that the ID does not change each time the configuration is
updated. IDs are tracked by the position of the processor block, so reordering inline processors moves
their IDs. Set `id` to keep a specific ID, such as the ID of a processor created in the Bindplane UI.

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
| `route_id`          | string       | required | A unique string that identifies the route. |
| `telemetry_type`    | enum         | required | The telemetry type to route. |
| `components`        | list(string) | required | One or more components to route the telemetry to. |

Telemetry types include the following
- `logs`
- `metrics`
- `traces`

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

### Advanced

| Option             | Type            | Default  | Description                                                                 |
| ------------------ | --------------- | -------- | --------------------------------------------------------------------------- |
| `metrics`          | block (single)  | optional | Metrics configuration options. See the [metrics section](#metrics) below.   |

#### Metrics

| Option             | Type            | Default  | Description                                                                 |
| ------------------ | --------------- | -------- | --------------------------------------------------------------------------- |
| `port`    | int             | optional | The port used for telemetry. Requires Bindplane v1.90.2 or newer.            |
| `level`   | string          | optional | The level of detail for telemetry. Valid values are 'basic', 'detailed'. Requires Bindplane v1.92 or newer.    |

### Gateway Configurations

> [!NOTE]
> Gateway configurations require Bindplane v1.103.0 or newer.

Bindplane marks a configuration as a gateway configuration with the label `role=gateway`.
Gateway configurations run on the collectors of a [gateway](./bindplane_gateway.md), which receive
telemetry from edge collectors. Configurations without the label are edge configurations.

```hcl
resource "bindplane_configuration_v2" "gateway" {
  rollout  = true
  name     = "my-gateway-config"
  platform = "linux"

  labels = {
    role = "gateway"
  }

  # ... sources, destinations, etc.
}
```

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
      route_id = "host-metrics"
      telemetry_type = "metrics"
      components = [
        "destinations/${bindplane_destination.google.id}"
      ]
    }
  }

  source {
    name = bindplane_source.journald.name
    route {
      route_id = "journald-logs"
      telemetry_type = "logs"
      components = [
        "destinations/${bindplane_destination.google.id}"
      ]
    }
  }

  destination {
    name = bindplane_destination.google.name
    route_id = bindplane_destination.google.id
    processors = [
      bindplane_processor.batch.name
    ]
  }

  extensions = [
    bindplane_extension.pprof.name
  ]
  rollout_options {
    type = "progressive"
    parameters {
      name = "stages"
      value {
        labels = {
          env = "stage"
        }
        name = "stage"
      }
      value {
        labels = {
          env = "production"
        }
        name = "production"
      }
    }
  }
}
```

### Inline Components

This example defines a routing connector and a processor inline, instead of creating
[bindplane_connector](./bindplane_connector.md) and [bindplane_processor](./bindplane_processor.md) resources.
Logs from production are filtered by severity and batched by a processor group before they are sent to
the destination. All other logs are sent to the destination directly.

```hcl
resource "bindplane_configuration_v2" "inline" {
  rollout  = true
  name     = "my-inline-config"
  platform = "linux"

  source {
    name = bindplane_source.journald.name
    route {
      route_id       = "router"
      telemetry_type = "logs"
      components = [
        "connectors/router"
      ]
    }
  }

  connector {
    route_id = "router"
    type     = "routing"
    parameters_json = jsonencode([
      {
        name  = "telemetry_types"
        value = ["Logs"]
      },
      {
        name = "routes"
        value = [
          {
            id = "production"
            condition = {
              ottl = "attributes[\"env\"] == \"production\""
            }
          },
          {
            # A route without a condition receives all remaining telemetry
            id = "default"
          }
        ]
      }
    ])

    # Each route_id matches the id of an entry in the routes parameter
    route {
      route_id       = "production"
      telemetry_type = "logs"
      components = [
        "processors/production"
      ]
    }
    route {
      route_id       = "default"
      telemetry_type = "logs"
      components = [
        "destinations/${bindplane_destination.google.id}"
      ]
    }
  }

  processor_group {
    route_id = "production"
    parameters_json = jsonencode([
      {
        name  = "telemetry_types"
        value = ["Logs"]
      }
    ])

    processor {
      type         = "filter_severity"
      display_name = "Drop Debug Logs"
      parameters_json = jsonencode([
        {
          name  = "severity"
          value = "INFO"
        }
      ])
    }

    processor {
      name = bindplane_processor.batch.name
    }

    route {
      route_id       = "destination"
      telemetry_type = "logs"
      components = [
        "destinations/${bindplane_destination.google.id}"
      ]
    }
  }

  destination {
    route_id = bindplane_destination.google.id
    name     = bindplane_destination.google.name
  }
}
```
