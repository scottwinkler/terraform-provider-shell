# terraform-provider-shell
[![GitHub Actions](https://img.shields.io/endpoint.svg?url=https%3A%2F%2Factions-badge.atrox.dev%2Fatrox%2Fsync-dotenv%2Fbadge)](https://actions-badge.atrox.dev/atrox/sync-dotenv/goto)
## Introduction
This plugin is for wrapping shell scripts to make them fully fledged terraform resources. Please note that this is a backdoor into the terraform lifecycle management, so it is up to you to implement your resources properly. It is recommended that you at least have some familiarity with the internals of Terraform before attempting to use this provider. If you can't write your own provider from scratch then you probably shouldn't be using this.

## Prerequisites
Get some coffee! ☕

## Installing
To use this plugin, go to releases and download the binary for your specific OS and architecture. Then you will need to trim the name of the file to get rid of the suffix (e.g. terraform-provider-shell_v1.0.0.darwin_amd64 -> terraform-provider-shell_v1.0.0). This suffix is only used to help you identify which binary to download and will cause errors if left on. Finally, you can install this plugin by either putting it in your `~/.terraform/plugins` folder or in your terraform workspace and performing a "terraform init".

## Examples
There is nothing to configure for the provider, you can declare it like so (or even omit it entirely):

```
provider "shell" {}
```
To use a data resource you need to implement the read command. Any output to stdout or stderr will show up in the logs, but to save the state, you must output a JSON payload to stdout. The last JSON object printed to stdout will be taken to be the state. The JSON can be a complex nested JSON, but will be flattened into a `map[string]string`. The reason for this is that your JSON payload variables can be accessed from the output map of this resource and used like a normal terraform output, so the value must be a string.

```
data "shell_script" "user" {
	lifecycle_commands {
		read = <<-EOF
		  echo "{\"user\": \"$(whoami)\"}"
		EOF
	}
}

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

Another data source example, this time to get the weather in San Francisco might be:

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
Resources are a bit more complicated. At a minimum, you must implement the `CREATE`, and `DELETE` lifecycle commands. `READ` and `UPDATE` are optional arguments.

* If you choose not to implement the `READ` command, then `CREATE` (and `UPDATE` if you are using it) must output the state in the form of a properly formatted JSON. The local state will not be synced with the actual state, but for many applications that is not a problem.

* If you choose not to implement `UPDATE`, then if a change occurs that would trigger an update, the resource will be instead be destroyed and then recreated - same as `ForceNew`. Again, for many applications this is not a problem, as `UPDATE` can be tricky to use. 

I suggest starting off with just `CREATE` and `DELETE` and then implementing `READ` and `UPDATE` only if you really need it. If you choose to implement `READ`, then you must output the state in the form of a properly formatted JSON, and you should not output the state in either the create or update scripts (otherwise it will be overridden). See the examples in the test folder for how to do each of these.

A complete example that uses all four lifecycle commands is shown below:

	resource "shell_script" "test" {
		lifecycle_commands {
			create = file("${path.module}/scripts/create.sh")
			read   = file("${path.module}/scripts/read.sh")
			update = file("${path.module}/scripts/update.sh")
			delete = file("${path.module}/scripts/delete.sh")
		}

		working_directory = path.module

		environment = {
			yolo = "yolo"
			ball = "room"
		}
	}

	output "commit_id" {
		value = shell_script.test.output["commit_id"]
	}

In the example I am setting the `working_directory` argument (which switches the current working directory), some environment variables that will be utilized by all my scripts, and configuring my lifecycle commands for `CREATE`, `READ`, `UPDATE` and `DELETE`. `CREATE`and `UPDATE` will modify the resource but not update the state, while `READ` updates the state but does not modify the resource.

An example shell script resouce could have a file being written to in the `CREATE`. `READ` would simply cat that previously created file and output it to stdout. `UPDATE` could measure the changes from the old state (available through stdin) and the new state (available through environment variables) to decide how best to handle an update. Again since this is a custom resource it is up to you to decide how best to handle updates, in many cases it may make sense not to implement `UPDATE` at all and rely on just `CREATE`/`READ`/`DELETE`.

`DELETE` needs to clean up any resources that were created but does not need to return anything. State data is available in the `output` variable, which is mapped from the JSON of your read command.

Stdout and stderr are also available in the debug log files. You can get this by setting:

```
export TF_LOG=debug
```

### Interpreter
By default, the scripts will be executed by `cmd /C` on windows and `/bin/sh -c` on linux. You can overwrite this setting using `interpreter`:

	resource "shell_script" "test" {
		lifecycle_commands {
			create = file("create.sh")
			read   = file("read.sh")
			update = file("update.sh")
			delete = file("delete.sh")
		}

		interpreter = {
			shell = "/bin/bash"
			flag = "-c"
		}
	}

* `shell` (optional) contains the path to the shell binary. If empty or non present, it will be set to `cmd` on windows and `/bin/sh` otherwise.
* `flag` (optional) contains the options passed to the shell (this example will execute `/bin/bash -c create.sh`). Can be empty. 

## Python Support
There is now an example for how to use the shell provider to invoke python files. Please check in the test/python-example folder for more information on this. Essentially it is an adapter around the shell resource that invokes methods on an interface that you implement.

## Python and Golang Support
There is now an example for how to use the shell provider to invoke python and golang files. Please check in the `examples/python-adapter` and `examples/golang-adapter` folder for more information on this. Essentially it is an adapter around the `shell_resource` that invokes methods on an interface that you implement.

## Develop
If you wish to build this yourself, follow the instructions:

```
	cd terraform-provider-shell
	make all
```
