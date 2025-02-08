---
subcategory: "Pipeline"
description: |-
  A Connector creates a BindPlane OP connector that can be attached
  to a Configuration's sources or destinations.
---

# bindplane_connector

The `bindplane_connector` resource creates a BindPlane connector from a BindPlane
connector-type. The connector can be used by multiple [configurations](./bindplane_configuration.md).

## Options

| Option              | Type   | Default  | Description                  |
| ------------------- | -----  | -------- | ---------------------------- |
| `name`              | string | required | The connector name.             |
| `type`              | string | required | The connector type.             |
| `parameters_json`   | string | optional | The serialized JSON representation of the connector type's parameters. |
| `rollout`           | bool   | required | Whether or not updates to the connector should trigger an automatic rollout of any configuration that uses it. |

## Examples

### Routing Connector

This example shows the Routing connector type with two routes.

```hcl
resource "bindplane_connector" "routing" {
  rollout = true
  name = "my-routing"
  type = "routing"
  parameters_json = jsonencode(
    [
      {
        "name": "telemetry_types",
        "value": [
          "Logs"
        ]
      },
      {
        "name": "routes",
        "value": [
          {
            "condition": {
              "ottl": "(attributes[\"env\"] == \"prod\")",
              "ottlContext": "resource",
              "ui": {
                "operator": "",
                "statements": [
                  {
                    "key": "env",
                    "match": "resource",
                    "operator": "Equals",
                    "value": "prod"
                  }
                ]
              }
            },
            "id": "route-1"
          },
          {
            "condition": {
              "ottl": "(attributes[\"env\"] == \"dev\")",
              "ottlContext": "resource",
              "ui": {
                "operator": "",
                "statements": [
                  {
                    "key": "env",
                    "match": "resource",
                    "operator": "Equals",
                    "value": "dev"
                  }
                ]
              }
            },
            "id": "route-2"
          }
        ]
      }
    ] 
  )
}
```

## Usage

You can find available connector types with the `bindplane get connector-type` command:
```bash
NAME   	DISPLAY	VERSION
count  	Count  	1
routing	Routing	1
```

You can view an individual connector type's options with the `bindplane get connector-type <name> -o yaml` command:
```yaml
# bindplane get connector-type routing -o yaml
apiVersion: bindplane.observiq.com/v1
kind: ConnectorType
metadata:
    id: 01JKGZE5JPHEN60VHQ2VTFDBFM
    name: routing
    displayName: Routing
    description: Route telemetry based on conditions
    icon: /icons/connectors/routing.svg
    hash: 69487953ba8144a063f4d855f95f129789fcfcf368570f47a713625083b7abc7
    version: 1
    dateModified: 2025-02-07T14:50:54.294599087-05:00
    stability: beta
spec:
    parameters:
        - name: telemetry_types
          label: Choose Telemetry Type
          description: Telemetry Type for the routes
          required: true
          type: telemetrySelector
          validValues:
            - Logs
            - Metrics
            - Traces
          default: []
          options:
            gridColumns: 12
        - name: routes
          label: Routes
...
```

You can view the json representation of the connector type's options with the `-o json` flag combined with `jq`.
For example, `bindplane get connector-type routing -o json | jq .spec.parameters` produces the following:
```json
[
  {
    "name": "telemetry_types",
    "label": "Choose Telemetry Type",
    "description": "Telemetry Type for the routes",
    "required": true,
    "type": "telemetrySelector",
    "validValues": [
      "Logs",
      "Metrics",
      "Traces"
    ],
    "default": [],
    "options": {
      "gridColumns": 12
    }
  },
  {
    "name": "routes",
    "label": "Routes",
    "description": "Telemetry will be sent to the first route it matches based on the condition. If\nthere is no condition specified for a route, all remaining telemetry will be sent\nto that route.\n",
    "required": true,
    "type": "routes",
    "default": [
      {
        "id": "route-1"
      },
      {
        "id": "route-2"
      }
    ],
    "options": {
      "gridColumns": 12
    },
    "properties": {
      "addButtonText": "Add Route",
      "condition": true,
      "routeBase": "route"
    }
  }
]
```

## Import

When using the [terraform import command](https://developer.hashicorp.com/terraform/cli/commands/import),
connector can be imported. For example:

```bash
terraform import bindplane_connector.connector {{name}}
```
