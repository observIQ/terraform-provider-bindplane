---
page_title: "Provider: BindPlane Observability Pipelines"
description: |-
   The BindPlane provider is used to configure your BindPlane OP resources
---

# BindPlane OP Provider

The BindPlane provider is used to configure your [BindPlane OP](https://observiq.com/solutions/bindplane-op/) resources.

To learn the basics of Terraform using this provider, follow the hands-on
[get started tutorials](https://developer.hashicorp.com/terraform/tutorials/gcp-get-started/infrastructure-as-code).

## Example Usage

A typical provider configuration will look something like:

```hcl
provider "bindplane" {
  remote_url = "http://localhost:3001"
  username = "admin"
  password = "admin"
}
```

If you have configuration questions, or general questions about using the provider, try checking out:

* The [BindPlane OP Community Slack](https://launchpass.com/bindplane)

## Releases

Interested in the provider's latest features, or want to make sure you're up to date?
Check out the [`bindplane` provider Releases](https://github.com/observIQ/terraform-provider-bindplane/releases)
and the [`bindplane-enterprise` provider Releases](https://github.com/observIQ/terraform-provider-bindplane-enterprise/releases)
for release notes and additional information.

