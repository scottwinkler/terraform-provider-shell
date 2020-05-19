package shell

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/mutexkv"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

// Provider returns a terraform.ResourceProvider.
func Provider() terraform.ResourceProvider {

	// The actual provider
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"environment": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     schema.TypeString,
			},
			"sensitive_environment": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     schema.TypeString,
			},
		},

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
	config := Config{
		Environment:          d.Get("environment").(map[string]interface{}),
		SensitiveEnvironment: d.Get("sensitive_environment").(map[string]interface{}),
	}
	return config.Client()
}

// This is a global MutexKV for use within this plugin.
var shellMutexKV = mutexkv.NewMutexKV()

const shellScriptMutexKey = "shellScriptMutexKey"
