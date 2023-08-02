---
page_title: "Provider: BindPlane Observability Pipelines"
description: |-
   The BindPlane provider is used to configure your BindPlane OP resources
---

# BindPlane OP Provider

The BindPlane provider is used to configure your [BindPlane OP](https://observiq.com/solutions/bindplane-op/) resources.

To learn the basics of Terraform using follow the hands-on
[get started tutorials](https://developer.hashicorp.com/terraform/tutorials/gcp-get-started/infrastructure-as-code).

## Provider Configuration

The provider can be configured with options and environment variables.

| Option                      | Evironment                | Description                  |
| --------------------------- | ------------------------- | ---------------------------- |
| `remote_url`                | `BINDPLANE_TF_REMOTE_URL` | The URL for the BindPlane server.  |
| `username`                  | `BINDPLANE_TF_USERNAME`   | The BindPlane basic auth username. |
| `password`                  | `BINDPLANE_TF_PASSWORD`   | The BindPlane basic auth password. |
| `tls_certificate_authority` | `BINDPLANE_TF_TLS_CA`     | Path to x509 PEM encoded certificate authority to trust when connecting to BindPlane. |
| `tls_certificate`           | `BINDPLANE_TF_TLS_CERT`   | Path to x509 PEM encoded client certificate to use when mTLS is desired. |
| `tls_private_key`           | `BINDPLANE_TF_TLS_KEY`    | Path to x509 PEM encoded private key to use when mTLS is desired. |

### Enterprise Options

The [bindplane-enterprise](https://registry.terraform.io/providers/observIQ/bindplane-enterprise/latest) provider
extends the BindPlane OP provider with additional configuration options. The following options are specific to the
enterprise provider.

| Option                      | Evironment                | Description                  |
| --------------------------- | ------------------------- | ---------------------------- |
| `api_key`                   | `BINDPLANE_TF_API_KEY`    | The API key to use for authentication as an alternative to `username` and `password`. |

## Example Usage

### Basic Auth

Basic auth can be configured by setting `username` and `password` options or
by setting the `BINDPLANE_TF_USERNAME` and `BINDPLANE_TF_PASSWORD` environment
variables.

```hcl
provider "bindplane" {
  remote_url = "http://192.168.1.10:3001"
  username = "admin"
  password = "admin"
}
```

```hcl
// Assumes the BINDPLANE_TF_USERNAME and BINDPLANE_TF_PASSWORD
// environment variables are set.
provider "bindplane" {
  remote_url = "http://192.168.1.10:3001"
}
```

### TLS

```hcl
provider "bindplane" {
  remote_url = "https://192.168.1.10"
  tls_certificate_authority = "/opt/tls/bindplane-east1.crt"
}
```

### Enterprise

The enterprise provider can be configured by updating the `required_providers`
block.

```hcl
terraform {
  required_providers {
    bindplane = {
      source = "observiq/bindplane-enterprise"
    }
  }
}

provider "bindplane" {
  remote_url = "http://192.168.1.10:3001"
  api_key    = "xxx-xxx-xxx-xxx"
}
```

## Releases

Interested in the provider's latest features, or want to make sure you're up to date?
Check out the [`bindplane` provider Releases](https://github.com/observIQ/terraform-provider-bindplane/releases)
and the [`bindplane-enterprise` provider Releases](https://github.com/observIQ/terraform-provider-bindplane-enterprise/releases)
for release notes and additional information.

## External Links

* [BindPlane OP Docs](https://docs.bindplane.observiq.com/docs)
* [BindPlane OP Community Slack](https://launchpass.com/bindplane)
