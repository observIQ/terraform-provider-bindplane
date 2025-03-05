---
subcategory: "Pipeline"
description: |-
  A Source creates a Bindplane source that can be attached
  to a Configuration. Sources are used by agents to receive or
  scrape telemetry from a network or file.
---

# bindplane_source

The `bindplane_source` resource creates a source from a Bindplane
source-type. The source can be used by multiple [configurations](./bindplane_configuration.md).

## Options

| Option              | Type   | Default  | Description                  |
| ------------------- | -----  | -------- | ---------------------------- |
| `name`              | string | required | The source name.             |
| `type`              | string | required | The source type.             |
| `parameters_json`   | string | optional | The serialized JSON representation of the source type's parameters. |
| `rollout`           | bool   | required | Whether or not updates to the source should trigger an automatic rollout of any configuration that uses it. |

## Sensitive Values

See the [sensitive values](./sensitive_values.md) doc for details related to Terraform's handling
of sensitive parameters, such as passwords and API keys.

## Examples

### OTLP w/ Default Options

This example shows the [Open Telemetry](https://docs.bindplane.observiq.com/docs/opentelemetry) source type
with default parameters.

```hcl
resource "bindplane_source" "otlp" {
  rollout = true
  name = "my-otlp"
  type = "otlp"
}
```

### OTLP w/ Custom Parameters

This example shows the [Open Telemetry](https://docs.bindplane.observiq.com/docs/opentelemetry) source type
with custom parameters using the `parameters_json` option.

```hcl
resource "bindplane_source" "otlp" {
  rollout = true
  name = "my-otlp"
  type = "otlp"
  parameters_json = jsonencode(
    [
      {
        "name": "http_port",
        "value": 44314
      },
      {
        "name": "grpc_port",
        "value": 0
      }
    ]
  )
}
```

## Usage

You can find available source types with the `bindplane get source-type` command:
```bash
NAME             DISPLAY          VERSION 
aerospike        Aerospike        3      	
apache_combined  Apache Combined  1      	
apache_common    Apache Common    1      	
apache_http      Apache HTTP      1   
...
```

You can view an individual source type's options with the `bindplane get source-type <name> -o yaml` command:
```yaml
# bindplane get source-type otlp -o yaml
apiVersion: bindplane.observiq.com/v1
kind: SourceType
metadata:
    id: 01H4KKMG3D14D8QM6BME0DMEPE
    name: otlp
    displayName: OpenTelemetry (OTLP)
    description: Receive metrics, logs, and traces from OTLP exporters.
    version: 3
spec:
    version: 0.0.1
    parameters:
        - name: listen_address
          label: Listen Address
          description: The IP address to listen on.
          type: string
          default: 0.0.0.0
...
```

You can view the json representation of the source type's options with the `-o json` flag combined with `jq`.
For example, `bindplane get source-type otlp -o json | jq .spec.parameters` produces the following:
```json
[
  {
    "name": "listen_address",
    "label": "Listen Address",
    "description": "The IP address to listen on.",
    "type": "string",
    "default": "0.0.0.0",
    "options": {}
  },
  {
    "name": "grpc_port",
    "label": "GRPC Port",
    "description": "TCP port to receive OTLP telemetry using the gRPC protocol. The port used must not be the same as the HTTP port. Set to 0 to disable.",
    "type": "int",
    "default": 4317,
    "options": {}
  },
  {
    "name": "http_port",
    "label": "HTTP Port",
    "description": "TCP port to receive OTLP telemetry using the HTTP protocol. The port used must not be the same as the gRPC port. Set to 0 to disable.",
    "type": "int",
    "default": 4318,
    "options": {}
  }
]
```

Use the JSON output as a reference when writing the `bindplane_source` resource configuration. This example sets
the `http_port` and `grpc_port` for an `otlp` source.

```hcl
resource "bindplane_source" "otlp" {
  rollout = true
  name = "my-otlp"
  type = "otlp"
  parameters_json = jsonencode(
    [
      {
        "name": "http_port",
        "value": 44314
      },
      {
        "name": "grpc_port",
        "value": 0
      }
    ]
  )
}
```

After applying the configuration with `terraform apply`, you can view the source with
the `bindplane get source` commands.

```bash
NAME        	TYPE
my-otlp     	otlp:1
```
```yaml
# bindplane get source my-otlp -o yaml
apiVersion: bindplane.observiq.com/v1
kind: Source
metadata:
    id: 01H4KKQ2KGJW3T8A8VB4JS7VZ6
    name: my-otlp
    hash: 1e4e53cb713bbcc097af8315320a4c67cd7e86a5c46235aa0afe0ce7c1631af2
    version: 1
    dateModified: 2023-07-06T15:17:08.060680291-04:00
spec:
    type: otlp:3
    parameters:
        - name: http_port
          value: 44314
        - name: grpc_port
          value: 0
status:
    latest: true
```

## Import

When using the [terraform import command](https://developer.hashicorp.com/terraform/cli/commands/import),
source can be imported. For example:

```bash
terraform import bindplane_source.source {{name}}
```