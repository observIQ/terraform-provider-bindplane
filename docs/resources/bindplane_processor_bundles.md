---
subcategory: "Pipeline"
description: |-
  A Processor Bundle creates a BindPlane OP processor bundle that can be attached
  to a Configuration's sources or destinations.
---

# bindplane_processor

The `bindplane_processor_bundle` resource creates a [BindPlane Processor Bundle](https://bindplane.com/docs/feature-guides/processor-bundles)
The processor bundle can be used by multiple [configurations](./bindplane_configuration.md).

## Options

| Option              | Type   | Default  | Description                  |
| ------------------- | -----  | -------- | ---------------------------- |
| `name`              | string | required | The processor name.             |
| `processor`         | processor block | required | One or more processor blocks. |
| `rollout`           | bool   | required | Whether or not updates to the processor should trigger an automatic rollout of any configuration that uses it. |

Processor block supports the following:

| Option             | Type   | Default  | Description                  |
| ------------------ | -----  | -------- | ---------------------------- |
| `name`             | string | required | The name of the processor to include in the bundle. |

## Usage

This example shows how to combine the batch and json processors
into a processor bundle.

```hcl
resource "bindplane_processor" "json-parse-body" {
  rollout = false
  name = "json-parse-body"
  type = "parse_json"
  parameters_json = jsonencode(
    [
      {
        "name": "telemetry_types",
        "value": [
          "Logs",
        ]
      },
      {
        "name": "log_source_field_type",
        "value": "Body"
      },
      {
        "name": "log_body_source_field",
        "value": ""
      },
      {
        "name": "log_target_field_type",
        "value": "Body"
      }
    ]
  )
}

resource "bindplane_processor" "batch" {
  rollout = false
  name    = "example-batch"
  type    = "batch"
}

resource "bindplane_processor_bundle" "bundle" {
  rollout = true
  name = "my-bundle"

  processor {
    name = bindplane_processor.json-parse-body.name
  }

  processor {
    name = bindplane_processor.batch.name
  }
}
```

After applying the configuration with `terraform apply`, you can view the processor bundle with
the `bindplane get processor "my-bundle"` command.

```bash
NAME     	TYPE 
my-bundle	processor_bundle:1
```
```yaml
# bindplane get processor my-bundle -o yaml
apiVersion: bindplane.observiq.com/v1
kind: Processor
metadata:
    id: 01JKEX6ZZNHHNX171N8JKQC57M
    name: my-bundle
    hash: 5ee0be3c33158b0452bf77da9413c4a571c2b9c407a2b84481741067c7c962b8
    version: 1
    dateModified: 2025-02-06T19:33:33.191627322-05:00
spec:
    type: processor_bundle:1
    processors:
        - id: p-example-batch
          name: example-batch:1
        - id: p-json-parse-body
          name: json-parse-body:1
status:
    latest: true
```

## Import

When using the [terraform import command](https://developer.hashicorp.com/terraform/cli/commands/import),
processor bundles can be imported. For example:

```bash
terraform import bindplane_processor_bundle.bundle {{name}}
```
