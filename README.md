BindPlane Terraform Provider
==========================

[![CI](https://github.com/observIQ/terraform-provider-bindplane/actions/workflows/ci.yml/badge.svg)](https://github.com/observIQ/terraform-provider-bindplane/actions/workflows/ci.yml)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

Terraform provider for observIQ's agent management platform, [BindPlane OP](https://github.com/observIQ/bindplane-op).

## Usage

### Provider Configuration

The provider can be configured with options
and environment variables.

| Option                      | Evironment                | Description                  |
| --------------------------- | ------------------------- | ---------------------------- |
| `remote_url`                | `BINDPLANE_TF_REMOTE_URL` | The URL for the BindPlane server.  |
| `username`                  | `BINDPLANE_TF_USERNAME`   | The BindPlane basic auth username. |
| `password`                  | `BINDPLANE_TF_PASSWORD`   | The BindPlane basic auth password. |
| `tls_certificate_authority` | `BINDPLANE_TF_TLS_CA`     | Path to x509 PEM encoded certificate authority to trust when connecting to BindPlane. |
| `tls_certificate`           | `BINDPLANE_TF_TLS_CERT`   | Path to x509 PEM encoded client certificate to use when mTLS is desired. |
| `tls_private_key`           | `BINDPLANE_TF_TLS_KEY`    | Path to x509 PEM encoded private key to use when mTLS is desired. |

#### Basic Auth

Basic auth can be configured by setting `username` and `password` options or
by setting the `BINDPLANE_TF_USERNAME` and `BINDPLANE_TF_PASSWORD` environment
variables.

```tf
provider "bindplane" {
  remote_url = "http://192.168.1.10:3001"
  username = "admin"
  password = "admin"
}
```

```tf
// Assumes the BINDPLANE_TF_USERNAME and BINDPLANE_TF_PASSWORD
// environment variables are set.
provider "bindplane" {
  remote_url = "http://192.168.1.10:3001"
}
```

#### TLS

```tf
provider "bindplane" {
  remote_url = "https://192.168.1.10"
  tls_certificate_authority = "/opt/tls/bindplane-east1.crt"
}
```

## Resource Documentation

See the [resource docs](./doc/resources/) for individual resource documentation.
