---
page_title: "Provider: Bindplane Observability Pipelines"
description: |-
   The Bindplane provider is used to configure your Bindplane resources
---

# Bindplane Provider

The Bindplane provider is used to configure your [Bindplane](https://observiq.com/solutions/bindplane-op/) resources.

To learn the basics of Terraform using follow the hands-on
[get started tutorials](https://developer.hashicorp.com/terraform/tutorials/gcp-get-started/infrastructure-as-code).

## Provider Configuration

The provider can be configured with options and environment variables.

| Option                      | Evironment                | Description                  |
| --------------------------- | ------------------------- | ---------------------------- |
| `remote_url`                | `BINDPLANE_TF_REMOTE_URL` | The URL for the Bindplane server.  |
| `api_key`                   | `BINDPLANE_TF_API_KEY`    | The API key to use for authentication as an alternative to `username` and `password`. |
| `username`                  | `BINDPLANE_TF_USERNAME`   | The Bindplane basic auth username. |
| `password`                  | `BINDPLANE_TF_PASSWORD`   | The Bindplane basic auth password. |
| `tls_certificate_authority` | `BINDPLANE_TF_TLS_CA`     | Path to x509 PEM encoded certificate authority to trust when connecting to Bindplane. |
| `tls_certificate`           | `BINDPLANE_TF_TLS_CERT`   | Path to x509 PEM encoded client certificate to use when mTLS is desired. |
| `tls_private_key`           | `BINDPLANE_TF_TLS_KEY`    | Path to x509 PEM encoded private key to use when mTLS is desired. |

## Example Usage

### Basic Auth

Basic auth can be configured by setting `username` and `password` options or
by setting the `BINDPLANE_TF_USERNAME` and `BINDPLANE_TF_PASSWORD` environment
variables.

```hcl
terraform {
  required_providers {
    bindplane = {
      source = "observiq/bindplane"
    }
  }
}

provider "bindplane" {
  remote_url = "http://192.168.1.10:3001"
  username = "admin"
  password = "admin"
}
```

```hcl
terraform {
  required_providers {
    bindplane = {
      source = "observiq/bindplane"
    }
  }
}

// Assumes the BINDPLANE_TF_USERNAME and BINDPLANE_TF_PASSWORD
// environment variables are set.
provider "bindplane" {
  remote_url = "http://192.168.1.10:3001"
}
```

### TLS

```hcl
terraform {
  required_providers {
    bindplane = {
      source = "observiq/bindplane"
    }
  }
}

provider "bindplane" {
  remote_url = "https://192.168.1.10"
  tls_certificate_authority = "/opt/tls/bindplane-east1.crt"
}
```

### API Key

API key can be used to authenticate as an alternative to username and password.

```hcl
terraform {
  required_providers {
    bindplane = {
      source = "observiq/bindplane"
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
for release notes and additional information.

## External Links

* [Bindplane Docs](https://bindplane.com/docs)
* [Bindplane Community Slack](https://launchpass.com/bindplane)
