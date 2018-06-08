package shell

import (
	"github.com/hashicorp/terraform/helper/mutexkv"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

// Provider returns a terraform.ResourceProvider.
func Provider() terraform.ResourceProvider {

	// The actual provider
	return &schema.Provider{
		Schema: map[string]*schema.Schema{},

		DataSourcesMap: map[string]*schema.Resource{
			"shell_script": dataSourceShellScript(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"shell_script": resourceShellScript(),
		},
		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	config := Config{}

	return config.Client()
}

// This is a global MutexKV for use within this plugin.
var shellMutexKV = mutexkv.NewMutexKV()

const shellScriptMutexKey = "shellScriptMutexKey"
