---
layout: "shell"
page_title: "Provider: Shell"
sidebar_current: "docs-shell-index"
description: |-
  Terraform Provider Shell.
---

# Shell Provider

This plugin is for wrapping shell scripts to make them fully fledged terraform resources. Note that this is a backdoor into the Terraform runtime. You can do some pretty dangerous things with this and it is up to you to make sure you don't get in trouble.

Since this provider is rather different than most other provider, it is recommended that you at least have some familiarity with the internals of Terraform before attempting to use this provider.

Use the navigation to the left to read about the available resources.

## Example Usage

```hcl
provider "shell" {
  sensitive_environment = {
    OAUTH_TOKEN = var.oauth_token
  }
}

resource "shell_script" "github_repository" {
  lifecycle_commands {
    create = file("${path.module}/scripts/create.sh")
    read   = file("${path.module}/scripts/read.sh")
    update = file("${path.module}/scripts/update.sh")
    delete = file("${path.module}/scripts/delete.sh")
  }

  environment = {
    NAME        = "HELLO-WORLD"
    DESCRIPTION = "description"
  }
}
```

Stdout and stderr stream to log files. You can get this by setting:

```
export TF_LOG=1
```
**Note:** if you are using sensitive_environment to set sensitive environment variables, these values won't show up in the logs