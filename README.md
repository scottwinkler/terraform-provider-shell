# terraform-provider-shell
[![GitHub Actions](https://img.shields.io/endpoint.svg?url=https%3A%2F%2Factions-badge.atrox.dev%2Fatrox%2Fsync-dotenv%2Fbadge)](https://actions-badge.atrox.dev/atrox/sync-dotenv/goto)
## Introduction
This plugin is for wrapping shell scripts to make them fully fledged terraform resources. Please note that this is a backdoor into the terraform life cycle management, so it is up to you to implement your resources properly. It is recommended that you at least have some familiarity with the internals of terraform before attempting to use this provider. If you can't write your own resource using lifecycle events then you probably shouldn't be using this.

## Prerequisites
Get some coffee! ☕

## Examples
There is nothing to configure for the provider, declare it like so

	provider "shell" {}

To use a data resource you need to implement the read command. Any output to stdout or stderr will show up in the logs, but to save the state, you must output to >&3. In addition, your output must be a properly formatted json object that can be marshalled into a map[string]string. The reason for this is that your json payload variables can be accessed from the output map of this resource and used like a normal terraform output.

	data "shell_script" "test1" {
		lifecycle_commands {
			read = <<EOF
			echo '{"commit_id": "b8f2b8b"}' >&3
			EOF
		}
	}

	#accessing the output from the data resource
	output "commit_id" {
  		value = data.shell_script.test1.output["commit_id"]
	}

Resources are a bit more complicated. You must implement the create, and delete lifecycle commands, but read and update are also optionally available. If you choose not to implement the read command, then create (and update if you are using it) must output the state in the form of a properly formatted json. The local state will not be synced with the actual state, but for many applications that is not a problem. If you choose not to implement update, then if a change occurs that would trigger an update the resource will be instead be destroyed and then recreated. Again, for many applications this is not a problem, update can be tricky to use as it depends a lot on the use case. If you implement read, then you must output the state in the form of a properly formatted json, and you should not output the state in either the create or update scripts. See the examples in the test folder for how to do each of these.

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

In the example above I am changing my working_directory, setting some environment variables that will be utilized by all my scripts, and configuring my lifecycle commands for create, read, update and delete. Create and Update should modify the resource but not update the state, while Read should update the state but not modify the resource. An example shell script resouce could have a file being written to in the create. Read would simply cat that previously created file and output it to >&3. Update could measure the changes from the old state (available through stdin) and the new state (implicitly available through environment variables) to decide how best to handle an update. Again since this is a custom resource it is up to you to decide how best to handle updates, in many cases it may make sense not to implement update at all and rely on just create/read/delete. Delete needs to clean up any resources that were created but does not need to return anything. State data is available in the output variable, which is mapped from the json of your read command.

Stdout and stderr are available in the debug log files. 

## Python Support
There is now an example for how to use the shell provider to invoke python files. Please check in the test/python-example folder for more information on this. Essentially it is an adapter around the shell resource that invokes methods on an interface that you implement.

## Develop
If you wish to build this yourself, follow the instructions:

	cd terraform-provider-shell
	go build
	
