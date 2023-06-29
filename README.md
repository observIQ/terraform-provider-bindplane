![Terraform](https://img.shields.io/badge/terraform-%235835CC.svg?style=for-the-badge&logo=terraform&logoColor=white) BindPlane Provider
==========================

[![CI](https://github.com/observIQ/terraform-provider-bindplane/actions/workflows/ci.yml/badge.svg)](https://github.com/observIQ/terraform-provider-bindplane/actions/workflows/ci.yml)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

Terraform provider for observIQ's agent management platform, [BindPlane](https://github.com/observIQ/bindplane).

## Usage

### Provider Configuration

The provider can be configured with options
and environment variables.

| Option                      | Evironment               | Default  | Description                  |
| --------------------------- | ------------------------ | -------- | ---------------------------- |
| `remote_url`                | `BINDPLANE_TF_REMOTE_URL` | required | The URL for the BindPlane server  |
| `username`                  | `BINDPLANE_TF_USERNAME`   | required | The BindPlane basic auth username |
| `password`                  | `BINDPLANE_TF_PASSWORD`   | required | The BindPlane basic auth password |
| `tls_certificate_authority` | `BINDPLANE_TF_TLS_CA`     | optional | Path to x509 PEM encoded certificate authority to trust when connecting to BindPlane |
| `tls_certificate`           | `BINDPLANE_TF_TLS_CERT`   | optional | Path to x509 PEM encoded client certificate to use when mTLS is desired |
| `tls_private_key`           | `BINDPLANE_TF_TLS_KEY`    | optional | Path to x509 PEM encoded private key to use when mTLS is desired |

**Configure with options:**

```tf
provider "bindplane" {
  remote_url = "http://192.168.1.10:3001"
  username = "admin"
  password = "admin"
}
```

**Configure with options and environment variables:**

```tf
// Assumes the BINDPLANE_TF_USERNAME and BINDPLANE_TF_PASSWORD
// environment variables are set.
provider "bindplane" {
  remote_url = "http://192.168.1.10:3001"
}
```

**Configure TLS:**

```tf
provider "bindplane" {
  remote_url = "https://192.168.1.10"
  tls_certificate_authority = "/opt/tls/bindplane-east1.crt"
}
```

**Configure Mutual TLS:**

Authentication with TLS can be achieved by setting the certificate authority,
client certificate, and private key.

```tf
provider "bindplane" {
  remote_url = "https://192.168.1.10"
  tls_certificate_authority = "/opt/tls/bindplane-east1.crt"
  tls_certificate = "/opt/tls/bindplane-client.crt"
  tls_private_key = "/opt/tls/bindplane-client.key"
}
```

### Resources

#### bindplane_raw_configuration

The `bindplane_raw_configuration` resource takes a raw OpenTelemetry configuration and applys it
to BindPlane. If the `match_labels` match an agent, BindPlane will push the configuration to the agent(s).

| Option              | Type   | Default  | Description                  |
| ------------------- | -----  | -------- | ---------------------------- |
| `name`              | string | required | The configuration name                                         |
| `labels`            | map    | required | Friendly labels                                                |
| `match_labels`      | map    | required | The labels that will be used for matching agents               |
| `raw_configuration` | string | required | The OpenTelemetry configuration that will be applied to agents |

The following example will match agents with labels `env=stage,platform=frontend`.

```tf
resource "bindplane_raw_configuration" "config" {
  name = "stage"
  labels = {
    env = "stage"
  }
  match_labels = {
    env = "stage"
    platform = "frontend"
  }
  raw_configuration = <<EOT
receivers:
  hostmetrics:
    collection_interval: 60s
    scrapers:
      cpu:
exporters:
  logging:
service:
  pipelines:
    metrics:
      receivers: [hostmetrics]
      exporters: [logging]
EOT
}
```
