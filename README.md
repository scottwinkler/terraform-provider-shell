Terraform Provider Shell
==================
https://registry.terraform.io/providers/scottwinkler/shell

This plugin is for wrapping shell scripts to make them fully fledged terraform resources. Note that this is a backdoor into the Terraform runtime. You can do some pretty dangerous things with this and it is up to you to make sure you don't get in trouble.

Since this provider is rather different than most other provider, it is recommended that you at least have some familiarity with the internals of Terraform before attempting to use this provider.

**Note:** many people use this provider for wrapping APIs of resources that are not supported by existing providers. For an example of using this provider to manage a Github repo resource, see `examples/github-repo`

Requirements
------------

-	[Terraform](https://www.terraform.io/downloads.html) >= 0.12.x
-	[Go](https://golang.org/doc/install) >= 1.12

Building The Provider
---------------------

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the Go `install` command: 
```sh
$ go install
```

Adding Dependencies
---------------------

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using Go modules.

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.


Using the provider
----------------------

You can use this provider to make custom external resources and data sources:

```
terraform {
  required_providers {
    shell = {
      source = "scottwinkler/shell"
      version = "1.7.7"
    }
  }
}

variable "oauth_token" {
  type = string
}

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

Developing the Provider
---------------------------

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```sh
$ make testacc
```
