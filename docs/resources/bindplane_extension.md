---
subcategory: "Pipeline"
description: |-
  An Extension creates a BindPlane OP extension that can be attached
  to a Configuration's sources or destinations.
---

# bindplane_extension

The `bindplane_extension` resource creates a BindPlane extension from a BindPlane
extension-type. The extension can be used by multiple [configurations](./bindplane_configuration.md).

## Options

| Option              | Type   | Default  | Description                  |
| ------------------- | -----  | -------- | ---------------------------- |
| `name`              | string | required | The extension name.          |
| `type`              | string | required | The extension type.          |
| `parameters_json`   | string | optional | The serialized JSON representation of the extension type's parameters. |
| `rollout`           | bool   | required | Whether or not updates to the extension should trigger an automatic rollout of any configuration that uses it. |

## Sensitive Values

See the [sensitive values](./sensitive_values.md) doc for details related to Terraform's handling
of sensitive parameters, such as passwords and API keys.

## Examples

### Go pprof w/ Default Options

This example shows the [pprof](https://observiq.com/docs/agent-configuration/extensions/pprof) extension type
with default parameters.

```hcl
resource "bindplane_extension" "pprof" {
  rollout = true
  name = "my-pprof"
  type = "pprof"
}
```

### Health Check w/ Custom Parameters

This example shows the [health_check](https://observiq.com/docs/agent-configuration/extensions/health_check) extension type
with custom parameters using the `parameters_json` option.

```hcl
resource "bindplane_extension" "health_check" {
  rollout = true
  name = "my-healthcheck"
  type = "health_check"
  parameters_json = jsonencode(
    [
      {
        "name": "listen_address",
        "value": "0.0.0.0"
      },
      {
        "name": "listen_port",
        "value": 8080,
      },
    ]
  )
}
```

## Usage

You can find available extension types with the `bindplane get extension-type` command:
```bash
NAME        	DISPLAY                        	VERSION 
custom      	Custom                         	1      	
health_check	Health Check                   	2      	
pprof       	Go Performance Profiler (pprof)	2   	
```

You can view an individual extension type's options with the `bindplane get extension-type <name> -o yaml` command:
```yaml
# bindplane get extension-type pprof -o yaml
apiVersion: bindplane.observiq.com/v1
kind: ExtensionType
metadata:
    id: 01HMC34DS83BXCV3ZACZB8EA1D
    name: pprof
    displayName: Go Performance Profiler (pprof)
    description: Enable the Go performance profiler.
    labels:
        category: Advanced
    version: 2
spec:
    version: 1.0.1
    parameters:
        - name: listen_address
          label: Listen Address
          description: The IP address or hostname to bind the profiler to.  Set to 0.0.0.0 to listen on all network interfaces.
          required: true
          type: string
          default: 127.0.0.1
        - name: tcp_port
          label: Port
          description: The TCP port to bind the profiler to.
          required: true
          type: int
          default: 1777
...
```

You can view the json representation of the extension type's options with the `-o json` flag combined with `jq`.
For example, `bindplane get extension-type pprof -o json | jq .spec.parameters` produces the following:
```json
[
  {
    "name": "listen_address",
    "label": "Listen Address",
    "description": "The IP address or hostname to bind the profiler to.  Set to 0.0.0.0 to listen on all network interfaces.",
    "required": true,
    "type": "string",
    "default": "127.0.0.1",
    "options": {}
  },
  {
    "name": "tcp_port",
    "label": "Port",
    "description": "The TCP port to bind the profiler to.",
    "required": true,
    "type": "int",
    "default": 1777,
    "options": {}
  },
  {
    "name": "block_profile_fraction",
    "label": "Block Profile Fraction",
    "description": "The fraction of blocking events that are profiled.  Must be between 0 and 1.",
    "type": "fraction",
    "default": 0,
    "advancedConfig": true,
    "options": {}
  },
  {
    "name": "mutex_profile_fraction",
    "label": "Mutex Profile Fraction",
    "description": "The fraction of mutex contention events that are profiled.  Must be between 0 and 1.",
    "type": "fraction",
    "default": 0,
    "advancedConfig": true,
    "options": {}
  },
  {
    "name": "should_write_file",
    "label": "Write CPU Profile to File",
    "description": "Whether or not to write the CPU profile to a file.",
    "type": "bool",
    "default": false,
    "advancedConfig": true,
    "options": {
      "sectionHeader": true
    }
  },
  {
    "name": "cpu_profile_file_name",
    "label": "CPU Profile File Name",
    "description": "The file name to write the CPU profile to.",
    "required": true,
    "type": "string",
    "default": "$OIQ_OTEL_COLLECTOR_HOME/observiq-otel-collector.pprof",
    "relevantIf": [
      {
        "name": "should_write_file",
        "operator": "equals",
        "value": true
      }
    ],
    "advancedConfig": true,
    "options": {}
  }
]
```

Use the JSON output as a reference when writing the `bindplane_extension` resource configuration. This example sets
the `listen_address`, and `listen_port` for a `pprof` extension.

```hcl
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
      },
    ]
  )
}
```

After applying the configuration with `terraform apply`, you can view the extension with
the `bindplane get extension` commands.

```bash
NAME    	TYPE   	DESCRIPTION 
my-pprof	pprof:1	  
```
```yaml
# bindplane get extension my-pprof -o yaml
apiVersion: bindplane.observiq.com/v1
kind: Extension
metadata:
    id: 01HQNNG5JFCY74WQ4MAEVD61H7
    name: my-pprof
    hash: c83267fa3be1323c7733c4db79dba0c840eb4de91af3def6abc9103fc0ff2415
    version: 1
    dateModified: 2024-02-27T11:13:55.15176503-05:00
spec:
    type: pprof:1
    parameters:
        - name: listen_address
          value: 0.0.0.0
        - name: tcp_port
          value: 5000
status:
    latest: true
```
