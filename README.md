# terraform-provider-shell
## Introduction
This plugin is for wrapping shell scripts to make them fully fledged terraform resources. It invokes shell scripts for each of the relevent lifecycle events: read, create, delete, and the output is saved in either stdout or stderr. Also if your stdout happens to be a json payload, then it will get mapped to an output map[string]string for ease of use within terraform. 

Note that this is essentially a backdoor into the terraform life cycle management, so it is up to you to implement your resources properly. It is recommended that you have at least have some familiarity with the internals of terraform before attempting to use this provider. No documentation is provided because if you can't figure out how to use this by reading the code and test cases then you probably shouldn't be using this.

This was inspired by: https://github.com/toddnni/terraform-provider-shell

## Prerequisites
Get some coffee! ☕

## Develop
If you wish to build this yourself, follow the instructions:

	cd ~/go/src/github.com/scottwinkler
	git clone http://github.com/scottwinkler/terraform-provider-shell.git
	cd terraform-provider-shell
	go get				
	go install
	mv ../../../../bin/terraform-provider-shell .			
	
