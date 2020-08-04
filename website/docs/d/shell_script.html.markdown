---
layout: "shell"
page_title: "Shell: shell_script"
sidebar_current: "docs-shell-data-source"
description: |-
  Shell script custom data source
---

# shell_script

The simplest example is the data source which implements only `Read()`. Any output to stdout or stderr will show up in the logs, but to save state, you must output a JSON payload to stdout. The last JSON object printed to stdout will be taken to be the output state. The JSON can be a complex nested JSON, but will be flattened into a `map[string]string`. The reason for this is that your JSON payload variables can be accessed from the output map of this resource and used like a normal terraform output, so the value must be a string. You can use the built-in jsondecode() function to read nested JSON values if you really need to.

## Example Usage

```hcl
data "shell_script" "user" {
	lifecycle_commands {
		read = <<-EOF
		  echo "{\"user\": \"$(whoami)\"}"
		EOF
	}
}
# "user" can be accessed like a normal Terraform map
output "user" {
	value = data.shell_script.user.output["user"]
}

data "shell_script" "weather" {
  lifecycle_commands {
    read = <<-EOF
        echo "{\"SanFrancisco\": \"$(curl wttr.in/SanFrancisco?format="%l:+%c+%t")\"}"
    EOF
  }
}

# value is: "SanFrancisco: ⛅️ +54°F"
output "weather" {
  value = data.shell_script.weather.output["SanFrancisco"]
}
```

## Attributes Reference

* `output` - A map of outputs
