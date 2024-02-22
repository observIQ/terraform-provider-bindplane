---
page_title: "Common issues/FAQ"
description: |-
  Common issues and frequently asked questions when using the provider.
---

# Google Provider Common Issues/FAQ

## Sensitive Values

The provider cannot detect changes to sensitive values (credentials). If a change to the sensitive value is made outside of Terraform, Terraform will
not detect configuration drift.

See the [Sensitive Values](/docs/guides/sensitive_values.md) documentation for more information.

## Dependent Resources Error

When removing a component from a configuration, Terraform may attempt to delete the component
before updating the configuration. This can cause the following error:

```
╷
│ Error: error while deleting destination with name example-custom: Dependent resources:
│ Configuration my-config
```

This can be prevented by using the [lifecycle Meta-Argument](https://developer.hashicorp.com/terraform/language/meta-arguments/lifecycle)
[create_before_destroy](https://developer.hashicorp.com/terraform/language/meta-arguments/lifecycle#create_before_destroy).

```tf
lifecycle {
  create_before_destroy = true
}
```

The create before destroy option will instruct Terraform to handle configuration updates before component deletion. This will result
in the following order of operations:

1. Remove the component from the configuration
2. Delete the component


```tf
resource "bindplane_configuration" "simple" {
  lifecycle {
    create_before_destroy = true
  }

  rollout = false
  name = "example-configuration-simple"
  platform = "linux"

  source {
    name = bindplane_source.host-custom.name
  }

  destination {
    name = bindplane_destination.grafana.name
  }
}
```
