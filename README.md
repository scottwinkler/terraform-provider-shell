# terraform-provider-shell
![Go](https://github.com/scottwinkler/terraform-provider-shell/workflows/Go/badge.svg)
## Introduction
This plugin is for wrapping shell scripts to make them fully fledged terraform resources. Note that this is a backdoor into the Terraform runtime. You can do some pretty dangerous things with this and it is up to you to make sure you don't get in trouble.

Since this provider is rather different than most other provider, it is recommended that you at least have some familiarity with the internals of Terraform before attempting to use this provider.

**Note:** many people use this provider for wrapping APIs of resources that are not supported by existing providers. For an example of using this provider to manage a Github repo resource, see `examples/github-repo`

## Requirements

-	[Terraform](https://www.terraform.io/downloads.html) 0.12.x
-	[Go](https://golang.org/doc/install) 1.13 (to build the provider plugin)

## Building The Provider

Clone repository to: `$GOPATH/src/github.com/scottwinkler/terraform-provider-shell`

```sh
$ mkdir -p $GOPATH/src/github.com/scottwinkler; cd $GOPATH/src/github.com/scottwinkler
$ git clone git@github.com:scottwinkler/terraform-provider-shell
```

Enter the provider directory and build the provider

```sh
$ cd $GOPATH/src/github.com/scottwinkler/terraform-provider-shell
$ make build
```

## Installing
To use this plugin, go to releases and download the binary for your specific OS and architecture. You can install the plugin by either putting it in your `~/.terraform/plugins` folder or in your terraform workspace by performing a `terraform init`.

## Configuring the Provider
There is nothing to configure for the provider, you can declare it like so (or omit entirely if you like):

```
provider "shell" {}
```

## Data Sources
The simplest example is the data source which implements only Read(). Any output to stdout or stderr will show up in the logs, but to save state, you must output a JSON payload to stdout. The last JSON object printed to stdout will be taken to be the output state. The JSON can be a complex nested JSON, but will be flattened into a `map[string]string`. The reason for this is that your JSON payload variables can be accessed from the output map of this resource and used like a normal terraform output, so the value must be a string. You can use jsondecode() to read nested JSON if you really need to.

Below is an example of using the data source. The output of `whoami` is stored in a JSON object for the key `user`

```
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
```

An apply would output the following:

```
shell_script.user: Creating...
shell_script.user: Creation complete after 0s [id=bpcs8j5grkris295e4qg]

Apply complete! Resources: 1 added, 0 changed, 0 destroyed.

Outputs:

user = swinkler
```
**Note:** the above example can be a very valuable way to get environment variables or other environment specific information into normal Terraform variables!

Another data source example, this time to get the weather in San Francisco:

```
data "shell_script" "weather" {
  lifecycle_commands {
    read = <<-EOF
        echo "{\"SanFrancisco\": \"$(curl wttr.in/SanFrancisco?format="%l:+%c+%t")\"}"
    EOF
  }
}

output "weather" {
  value = data.shell_script.weather.output["SanFrancisco"]
}
```

An apply would output the following:

```
shell_script.weather: Creating...
shell_script.weather: Creation complete after 0s [id=bpcs8j5grkris295e4qg]

Apply complete! Resources: 1 added, 0 changed, 0 destroyed.

Outputs:

weather = SanFrancisco: ⛅️ +54°F
```

## Resources
Resources are a bit more complicated. At a minimum, you must implement the `CREATE`, and `DELETE` lifecycle commands. `READ` and `UPDATE` are optional arguments.

* If you choose not to implement the `READ` command, then `CREATE` (and `UPDATE` if you are using it) must output JSON. The local state will not be synced with the actual state, but for many applications that is not a problem.

* If you choose not to implement `UPDATE`, then if a change occurs that would trigger an update, the resource will be instead be destroyed and then recreated - same as `ForceNew`. For many applications this is not a problem.

I suggest starting off with just `CREATE` and `DELETE` and then implementing `READ` and `UPDATE` only if you really need it. If you choose to implement `READ`, then you must output the state in the form of a properly formatted JSON, and you should not output the state in either the create or update scripts (otherwise it will be overridden). See the examples in the test folder for how to do each of these.

A complete example that uses all four lifecycle commands is shown below:

	resource "shell_script" "test" {
		lifecycle_commands {
			# i like to have scripts be in separate files if they are non-trivial
			create = file("${path.module}/scripts/create.sh")
			read   = file("${path.module}/scripts/read.sh")
			update = file("${path.module}/scripts/update.sh")
			delete = file("${path.module}/scripts/delete.sh")
		}

		# sets current working directory.
		working_directory = path.module

		# sets environment variables
		environment = {
			HELLO = "world"
		}

		# same as environment variables but won't be printed in logs or state
		sensitive_environment = {
			AWS_ACCESS_KEY_ID = var.aws_access_key
		}

		# triggers a force new update, like for null_resource
		triggers = {
			when_value_changed = var.some_value
		}
	}

	output "id" {
		value = shell_script.test.output["id"]
	}

Stdout and stderr stream to log files. You can get this by setting:

```
export TF_LOG=debug
```

## Testing
If you wish to run automated tests:

```sh
$ make test
```

