package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
	"github.com/scottwinkler/terraform-provider-shell/shell"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: shell.Provider})
}
