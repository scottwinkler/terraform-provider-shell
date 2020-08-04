---
layout: "shell"
page_title: "Shell: shell_script"
sidebar_current: "docs-shell-resource"
description: |-
  Shell script external resource
---

# shell_script

esources are a bit more complicated than data sources. At a minimum, you must implement the `CREATE`, and `DELETE` lifecycle commands. `READ` and `UPDATE` are optional arguments.

* If you choose not to implement the `READ` command, then `CREATE` (and `UPDATE` if you are using it) must output JSON. The local state will not be synced with the actual state, but for many applications that is not a problem.

* If you choose not to implement `UPDATE`, then if a change occurs that would trigger an update, the resource will be instead be destroyed and then recreated - same as `ForceNew`. For many applications this is not a problem.

I suggest starting off with just `CREATE` and `DELETE` and then implementing `READ` and `UPDATE` as needed. If you choose to implement `READ`, then you must output the state in the form of a properly formatted JSON, it should not alter the resource it is reading, and you should not output the state in either the create or update scripts (otherwise it will be overridden). See the examples in the test folder for how to do each of these.

## Example Usage

```hcl
variable "oauth_token" {
	type = string
}

provider "shell" {
	environment = {
		GO_PATH = "/Users/Admin/go"
	}
	sensitive_environment = {
		OAUTH_TOKEN = var.oauth_token
	}
	interpreter = ["/bin/sh", "-c"]
	enable_parallelism = false
}

resource "shell_script" "github_repository" {
	lifecycle_commands {
		//I suggest having these command be as separate files if they are non-trivial
		create = file("${path.module}/scripts/create.sh")
		read   = file("${path.module}/scripts/read.sh")
		update = file("${path.module}/scripts/update.sh")
		delete = file("${path.module}/scripts/delete.sh")
	}

	environment = {
		//changes to one of these will trigger an update
		NAME        = "HELLO-WORLD"
		DESCRIPTION = "description"
	}

	
	//sensitive environment variables are exactly the
	//same as environment variables except they don't
	//show up in log files
	sensitive_environment = {
		USERNAME = var.username
		PASSWORD = var.password
	}

	//this overrides the provider supplied interpreter
	//if you do not specify this then the default for your
	//machine will be used (/bin/sh for linux/mac and cmd for windows)
	interpreter = ["/bin/bash", "-c"]

	//sets current working directory
	working_directory = path.module

	//triggers a force new update if value changes, like null_resource
	triggers = {
		when_value_changed = var.some_value
	}
}

output "id" {
	value = shell_script.github_repository.output["id"]
}
```

## Argument Reference

The following arguments are supported:

* `output` - A map of outputs

