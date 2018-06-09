# terraform-provider-shell
## Introduction
This plugin is for wrapping shell scripts to make them fully fledged terraform resources. Please note that this is a backdoor into the terraform life cycle management, so it is up to you to implement your resources properly. It is recommended that you at least have some familiarity with the internals of terraform before attempting to use this provider. If you can't write your own resource using lifecycle events then you probably shouldn't be using this.

## Prerequisites
Get some coffee! ☕

## Examples
There is nothing to configure for the provider, declare it like so

	provider "shell" {}

To use a data resource you need to implement the read command. Stdout and stderr are available as outputs of this resource. In addition, if the output of your read command happens to be a json (map[string]string) then you can access these through the output map.

	data "shell_script" "test" {
		#kinda weird to have a map with only one variable but i wanted
		#to be consistent with the shell_script resource
		lifecycle_commands {
			read = <<EOF
			echo '{"commit_id": "b8f2b8b"}'
			EOF
		}
		working_directory = "."

		environment = {
			ydawgie = "scrubsauce"
		}
	}

	#accessing the output from the data resource
	output "commit_id" {
  		value = "${data.shell_script.test.output["commit_id"]}"
	}

Resources are a bit more complicated. You must implement the create, read and delete lifecycle commands. Update is not supported because that would be needlessly tricky. Instead, any change to this resource triggers it to be deleted and recreated, so you need to ensure that your delete and create commands are repeatable.

	resource "shell_script" "test" {
		lifecycle_commands {
			create = "bash create.sh"
			read   = "bash read.sh"
			delete = "bash delete.sh"
		}

		working_directory = "./scripts"

		environment = {
			yolo = "yolo"
		}
	}

In the example above I am changing my working_directory, setting some environment variables that will be utilized by all my scripts, and configuring my lifecycle commands for create, read and delete. Create and read should return the same output, but read doesn't create a new resource. An example shell script resouce could have a file being written to in the create, and then cat that file to stdout. Read would simply cat that previously created file to stdout. Delete needs to clean up any resources that were created. State data is available through stdin, which is the stdout of either the create or read lifecycle command.

Stdout, stderr and a variable map parsed from stdout (if the stdout happened to be a json payload) are available as outputs of this resource

## Develop
If you wish to build this yourself, follow the instructions:

	cd ~/go/src/githubdev.dco.elmae/CloudPlatform
	git clone http://githubdev.dco.elmae/CloudPlatform/terraform-provider-shell.git
	cd terraform-provider-shell
	go get				
	go install
	mv ../../../../bin/terraform-provider-shell .			
	
