# bindplane_destination

The `bindplane_destination` resource creates a BindPlane destination from a BindPlane
destination-type. The destination can be used by multiple [configurations](./bindplane_configuration.md).

## Options

| Option              | Type   | Default  | Description                  |
| ------------------- | -----  | -------- | ---------------------------- |
| `name`              | string | required | The destination name.             |
| `type`              | string | required | The destination type.             |
| `parameters_json`   | string | optional | The serialized JSON representation of the destination type's parameters. |
| `rollout`           | bool   | required | Whether or not updates to the destination should trigger an automatic rollout of any configuration that uses it. |

## Examples

### Google Cloud w/ Default Options

This example shows the [Google Cloud](https://docs.bindplane.observiq.com/docs/google-cloud) destination type
with default parameters.

```tf
resource "bindplane_destination" "googlecloud" {
  rollout = true
  name = "my-google"
  type = "googlecloud"
}
```

### Prometheus w/ Custom Parameters

This example shows the [Prometheus](https://docs.bindplane.observiq.com/docs/prometheus-1) destination type
with custom parameters using the `parameters_json` option.

```tf
resource "bindplane_destination" "prometheus" {
  rollout = true
  name = "my-prometheus"
  type = "prometheus"
  parameters_json = jsonencode(
    [
      {
        "name": "listen_address",
        "value": "0.0.0.0"
      },
      {
        "name": "listen_port",
        "value": 9000,
      },
      {
        "name": "namespace",
        "value": "otel"
      }
    ]
  )
}
```

## Usage

You can find available destination types with the `bindplane get destination-type` command:
```bash
NAME       DISPLAY    VERSION 
aws_s3     AWS S3     1      	
coralogix  Coralogix  3      	
custom     Custom     1      	
datadog    Datadog    3      	
...
```

You can view an individual destination type's options with the `bindplane get destination-type <name> -o yaml` command:
```yaml
# bindplane get destination-type prometheus -o yaml
apiVersion: bindplane.observiq.com/v1
kind: DestinationType
metadata:
    id: 01H4KKMFWK94WAEH7KAE7CCZTV
    name: prometheus
    displayName: Prometheus
    description: Serve Prometheus compatible metrics, scrapable by a Prometheus server.
    version: 1
spec:
    parameters:
        - name: listen_address
          label: Listen Address
          description: |
            The IP address the Prometheus exporter should  listen on, to be scraped by a Prometheus server.
          type: string
          default: 127.0.0.1
        - name: listen_port
          label: Listen Port
          description: |
            The TCP port the Prometheus exporter should listen on, to be scraped by a Prometheus server.
          type: int
          default: 9000
        - name: namespace
          label: Namespace
          description: When set, exports metrics under the provided value.
          type: string
          default: ""
          advancedConfig: true
...
```

You can view the json representation of the destination type's options with the `-o json` flag combined with `jq`.
For example, `bindplane get destination-type prometheus -o json | jq .spec.parameters` produces the following:
```json
[
  {
    "name": "listen_address",
    "label": "Listen Address",
    "description": "The IP address the Prometheus exporter should  listen on, to be scraped by a Prometheus server.\n",
    "type": "string",
    "default": "127.0.0.1",
    "options": {}
  },
  {
    "name": "listen_port",
    "label": "Listen Port",
    "description": "The TCP port the Prometheus exporter should listen on, to be scraped by a Prometheus server.\n",
    "type": "int",
    "default": 9000,
    "options": {}
  },
  {
    "name": "namespace",
    "label": "Namespace",
    "description": "When set, exports metrics under the provided value.",
    "type": "string",
    "default": "",
    "advancedConfig": true,
    "options": {}
  }
]
```

Use the JSON output as a reference when writing the `bindplane_destination` resource configuration. This example sets
the `listen_address`, `listen_port` and `namespace` for a `prometheus` destination.

```tf
resource "bindplane_destination" "prometheus" {
  rollout = true
  name = "my-prometheus"
  type = "prometheus"
  parameters_json = jsonencode(
    [
      {
        "name": "listen_address",
        "value": "0.0.0.0"
      },
      {
        "name": "listen_port",
        "value": 9000,
      },
      {
        "name": "namespace",
        "value": "otel"
      }
    ]
  )
}
```

After applying the configuration with `terraform apply`, you can view the destination with
the `bindplane get destination` commands.

```bash
NAME        	  TYPE
my-prometheus   prometheus:1
```
```yaml
# bindplane get destination my-prometheus -o yaml
apiVersion: bindplane.observiq.com/v1
kind: Destination
metadata:
    id: 01H4PCMDS67NMT0A02A7F675EF
    name: my-prometheus
    hash: 8b4f69b9bf78b7f289cf2bcdeccfd823fdefc193dfc0beeed1cb494466cc109d
    version: 1
    dateModified: 2023-07-06T15:59:57.222433696-04:00
spec:
    type: prometheus:1
    parameters:
        - name: listen_address
          value: 0.0.0.0
        - name: listen_port
          value: 9000
        - name: namespace
          value: otel
status:
    latest: true
```
