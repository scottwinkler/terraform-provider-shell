---
layout: "shell"
page_title: "Shell: shell_sensitive_script"
sidebar_current: "docs-shell-resource"
description: |-
  Shell script custom resource
---

# shell_script

The `shell_sensitive_script` resource shares the same interface as the `shell_script` resource, but its output is sensitive. As a result, the output will not be exposed in the logs.

## Example Usage

```hcl
resource "shell_sensitive_script" "special_secret" {
  lifecycle_commands {
    create = <<EOF
      set -e
      secret=$(openssl rand -hex 32)
      jq --null-input --arg secret "$secret" '{"value": $secret}'
    EOF
    delete = <<EOF
      # nothing
    EOF
  }
}
output "special_secret" {
  value     = shell_sensitive_script.special_secret.output["value"]
  sensitive = true
}
```

## Argument Reference

The following arguments are supported:

* `output` - A map of outputs
