package main

import (
	"github.com/hashicorp/terraform/plugin"
	"githubdev.dco.elmae/CloudPlatform/terraform-provider-shell/shell"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: shell.Provider})
}
