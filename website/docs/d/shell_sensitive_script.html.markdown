---
layout: "shell"
page_title: "Shell: shell_sensitive_script"
sidebar_current: "docs-shell-data-source"
description: |-
  Shell script custom data source
---

# shell_script

The `shell_sensitive_script` data shares the same interface as the `shell_script` data, but its output is sensitive. As a result, the output will not be exposed in the logs.


## Example Usage

```hcl
variable "token" {
  type = string
}

data "shell_sensitive_script" "secret" {
  lifecycle_commands {
    read = <<-EOF
      set -e
      secret=$(curl "https://example.com/secret" -H "Authorization: Basic $TOKEN")
      jq --null-input --arg secret "$secret" '{"value": $secret}'
    EOF
  }
  sensitive_environment = {
    TOKEN = var.token
  }
}

output "secret" {
  value = data.shell_sensitive_script.secret.output["value"]
  sensitive = true
}
```

## Attributes Reference

* `output` - A map of outputs
