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
| `profile`                   | `BINDPLANE_TF_PROFILE`    | The name of the bindplane profile to use. Profile options are overridden by other configured options. | 
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

#### Profile

A BindPlane profile can be used instead of specifying each option.


Asuming you have a profile named `local`, you can specify it in the provider configuration. This
example shows a profile with `username`, `password`, and `remoteURL` configured.
```bash
$ bindplane profile get local
name: local
apiVersion: bindplane.observiq.com/v1
auth:
  username: admin
  password: admin
network:
  remoteURL: http://localhost:3001

```
```tf
provider "bindplane" {
  profile = "local"
}
```

You can override options set by the profile by specifying them in the
provider configuration. This example shows that the `remote_url` can be overridden.
```tf
provider "bindplane" {
  profile = "local"
  remote_url = "https://bindplane.corp.net:443"
}
```

#### TLS

```tf
provider "bindplane" {
  remote_url = "https://192.168.1.10"
  tls_certificate_authority = "/opt/tls/bindplane-east1.crt"
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
