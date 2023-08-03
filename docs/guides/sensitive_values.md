---
page_title: "Sensitive Values"
description: |-
  BindPlane OP does not return sensitive values. Terraform practitioners should understand how the provider
  handles changes to sensitive values.
---

# Sensitive Values

Some BindPlane resources are configured with sensitive data, such as
passwords and API keys. 

It is important to understand:
- How BindPlane handles sensitive values
- How Terraform configuration and Git handle sensitive values
- How Terraform's state handles sensitive values

## BindPlane Sensitive Value Drift

BindPlane does not return the value for a sensitive parameter. This means
that Terraform will not detect changes to that value. It is important to
ensure that Terraform managed resources are not modified outside of Terraform
in order to ensure a consistent experience.

## Terraform Configuration and Git

When writing your Terraform configuration, it is possible to include
sensitive values that you might commit to Git. For example, the
[datadog destination](../../example/destination_datadog.tf) has an API Key
field.

```hcl
resource "bindplane_destination" "datadog" {
  rollout = true
  name = "example-datadog"
  type = "datadog"
  parameters_json = jsonencode(
    [
      {
        "name": "api_key",
        "value": "xxxx-xxxx-xxxx",
      },
    ]
  )
}
```

Use caution when saving sensitive value to configuration and git repositories.

## Terraform State

Terraform tracks the user's applied configuration in its state. This means that
even if your configuration code and git repositories are secure, the underlying
value is still human readable in the state backend.

```
➜  example git:(handle-sensitive-fields) ✗ terraform state show bindplane_destination.datadog
# bindplane_destination.datadog:
resource "bindplane_destination" "datadog" {
    id              = "01H6PD3AMDBW2Z1DCB917SCGFH"
    name            = "example-datadog"
    parameters_json = jsonencode(
        [
            {
                name  = "api_key"
                value = "xxxx-xxxx-xxxx"
            },
        ]
    )
    rollout         = true
    type            = "datadog"
}
```

See Hashicorp's [Sensitive Data in State](https://developer.hashicorp.com/terraform/language/state/sensitive-data)
documentation for more information, and how you can secure the state.
