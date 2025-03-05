---
subcategory: "Pipeline"
description: |-
  A Processor creates a Bindplane processor that can be attached
  to a Configuration's sources or destinations.
---

# bindplane_processor

The `bindplane_processor` resource creates a Bindplane processor from a Bindplane
processor-type. The processor can be used by multiple [configurations](./bindplane_configuration.md).

## Options

| Option              | Type   | Default  | Description                  |
| ------------------- | -----  | -------- | ---------------------------- |
| `name`              | string | required | The processor name.             |
| `type`              | string | required | The processor type.             |
| `parameters_json`   | string | optional | The serialized JSON representation of the processor type's parameters. |
| `rollout`           | bool   | required | Whether or not updates to the processor should trigger an automatic rollout of any configuration that uses it. |

## Sensitive Values

See the [sensitive values](./sensitive_values.md) doc for details related to Terraform's handling
of sensitive parameters, such as passwords and API keys.

## Examples

### Prometheus w/ Default Options

This example shows the [Batch](https://docs.bindplane.observiq.com/docs/batch) processor type
with default parameters.

```hcl
resource "bindplane_processor" "batch" {
  rollout = true
  name = "my-batch"
  type = "batch"
}
```

## Usage

You can find available processor types with the `bindplane get processor-type` command:
```bash
NAME               	DISPLAY              	VERSION 
batch              	Batch                	1      	
count_telemetry    	Count Telemetry      	1      	
custom             	Custom               	1       	
...
```

You can view an individual processor type's options with the `bindplane get processor-type <name> -o yaml` command:
```yaml
# bindplane get processor-type batch -o yaml
apiVersion: bindplane.observiq.com/v1
kind: ProcessorType
metadata:
    id: 01H4KKMG6RQ2SDM744P1255KC4
    name: batch
    displayName: Batch
    description: The batch processor accepts spans, metrics, or logs and places them into batches. Batching helps better compress the data and reduce the number of outgoing connections required to transmit the data. This processor supports both size and time based batching.
    labels:
        category: Advanced
    version: 1
spec:
    version: 0.0.1
    parameters:
        - name: send_batch_size
          label: Send Batch Size
          description: Number of spans, metric data points, or log records after which a batch will be sent regardless of the timeout.
          required: true
          type: int
          default: 8192
...
```

You can view the json representation of the processor type's options with the `-o json` flag combined with `jq`.
For example, `bindplane get processor-type batch -o json | jq .spec.parameters` produces the following:
```json
[
  {
    "name": "send_batch_size",
    "label": "Send Batch Size",
    "description": "Number of spans, metric data points, or log records after which a batch will be sent regardless of the timeout.",
    "required": true,
    "type": "int",
    "default": 8192,
    "options": {}
  },
  {
    "name": "send_batch_max_size",
    "label": "Send Batch Max Size",
    "description": "The upper limit of the batch size. 0 means no upper limit of the batch size. This property ensures that larger batches are split into smaller units. It must be greater than or equal to send batch size.",
    "required": true,
    "type": "int",
    "default": 0,
    "options": {}
  },
  {
    "name": "timeout",
    "label": "Timeout",
    "description": "Time duration after which a batch will be sent regardless of size. Example: 2s (two seconds)",
    "required": true,
    "type": "string",
    "default": "200ms",
    "options": {}
  }
]
```

Use the JSON output as a reference when writing the `bindplane_processor` resource configuration. This example sets
the `send_batch_size`, `send_batch_max_size` and `timeout` for a `batch` processor.

```hcl
resource "bindplane_processor" "batch" {
  rollout = true
  name = "my-batch"
  type = "batch"
  parameters_json = jsonencode(
    [
      {
        "name": "send_batch_size",
        "value": 200
      },
      {
        "name": "send_batch_max_size",
        "value": 400
      },
      {
        "name": "timeout",
        "value": "2s"
      }
    ]
  )
}
```

After applying the configuration with `terraform apply`, you can view the processor with
the `bindplane get processor` commands.

```bash
NAME        	  TYPE
my-batch      	batch:1 
```
```yaml
# bindplane get processor my-batch -o yaml
apiVersion: bindplane.observiq.com/v1
kind: Processor
metadata:
    id: 01H4PD71Z7TFJADC8N1MP1GPJC
    name: my-batch
    hash: c989d8c8a28b8a0083c3e423bbab539120b50be67a82966477e3b8b9942762dc
    version: 1
    dateModified: 2023-07-06T16:10:07.719917812-04:00
spec:
    type: batch:1
    parameters:
        - name: send_batch_size
          value: 200
        - name: send_batch_max_size
          value: 400
        - name: timeout
          value: 2s
status:
    latest: true

```

## Import

When using the [terraform import command](https://developer.hashicorp.com/terraform/cli/commands/import),
processor can be imported. For example:

```bash
terraform import bindplane_processor.processor {{name}}
```
