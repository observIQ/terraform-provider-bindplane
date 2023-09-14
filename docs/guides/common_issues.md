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
