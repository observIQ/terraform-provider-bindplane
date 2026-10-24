---
subcategory: "Fleet Management"
description: |-
  A Gateway groups the collectors of one gateway deployment, such as a group of
  VMs behind a load balancer, and assigns them a gateway configuration.
---

> [!NOTE]
> bindplane_gateway resources are supported by Bindplane version 1.103.0 and later.

# bindplane_gateway

The `bindplane_gateway` resource creates a Bindplane gateway. A gateway groups the collectors that make up
one gateway deployment, such as a group of VMs behind a load balancer. Bindplane lists these collectors
as gateway instances, not edge collectors, and assigns them the gateway's [configuration](./bindplane_configuration_v2.md).

Only manual mode gateways are supported. Gateways deployed by the Bindplane Kubernetes operator
(operator mode) are managed by the operator and cannot be read or imported by this resource.

## Gateway Membership

Collectors join a gateway with the label `gateway=<name>`, where `<name>` is the gateway's `name`.
The label is set when the collector is installed, not by Terraform. For example, with the
Bindplane collector install script, add `gateway=my-gateway` to the collector's labels.

## Options

| Option                   | Type   | Default  | Description                                                                                      |
| ------------------------ | ------ | -------- | ------------------------------------------------------------------------------------------------ |
| `name`                   | string | required | The resource name for the gateway, and the value of the collectors' `gateway` label. Cannot be changed after creation. |
| `display_name`           | string | optional | A user-friendly name for the gateway. Can be changed anytime.                                    |
| `description`            | string | optional | A description of the gateway.                                                                    |
| `configuration`          | string | optional | Name of the configuration assigned to the gateway's collectors. Must not include a version, such as `my-config:2`. |
| `load_balancer_hostname` | string | optional | Hostname of the load balancer in front of the gateway's collectors, where edge collectors send telemetry. |

The configuration should be a gateway configuration, labeled `role = "gateway"`. See
[gateway configurations](./bindplane_configuration_v2.md#gateway-configurations).

## Examples

This example creates a gateway configuration and a gateway that uses it.

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

resource "bindplane_gateway" "gateway" {
  name                   = "my-gateway"
  display_name           = "My Gateway"
  description            = "Gateway collectors behind a load balancer"
  configuration          = bindplane_configuration_v2.gateway.name
  load_balancer_hostname = "gateway.example.com"
}
```

## Configuration Dependency

When the configuration is referenced by its name attribute (e.g., `bindplane_configuration_v2.gateway.name`),
Terraform creates the configuration before the gateway.

If the configuration name is a string literal, the provider checks that the configuration exists when
the gateway is created and returns an error if it does not.

## Deleting a Gateway

Bindplane does not delete a gateway while collectors belong to it. Remove the collectors from
the gateway before running `terraform destroy`.

## Using Gateways

After applying the configuration with `terraform apply`, you can view the gateway with the
`bindplane get gateway` command:

```yaml
# bindplane get gateway my-gateway -o yaml
apiVersion: bindplane.observiq.com/v1
kind: Gateway
metadata:
    id: my-gateway
    name: my-gateway
    displayName: My Gateway
    description: Gateway collectors behind a load balancer
spec:
    configuration: my-gateway-config
    loadBalancerHostname: gateway.example.com
```

## Import

When using the [terraform import command](https://developer.hashicorp.com/terraform/cli/commands/import),
gateways can be imported by name. For example:

```bash
terraform import bindplane_gateway.gateway my-gateway
```
