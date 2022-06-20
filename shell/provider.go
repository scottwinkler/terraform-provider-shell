package shell

import (
	"sync"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Provider() *schema.Provider {

	// The actual provider
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"environment": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     schema.TypeString,
			},
			"sensitive_environment": {
				Type:      schema.TypeMap,
				Optional:  true,
				Sensitive: true,
				Elem:      schema.TypeString,
			},
			"interpreter": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"enable_parallelism": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
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
	var environment map[string]interface{}
	if v, ok := d.GetOk("environment"); ok {
		environment = v.(map[string]interface{})
	}
	var sensitiveEnvironment map[string]interface{}
	if v, ok := d.GetOk("sensitive_environment"); ok {
		sensitiveEnvironment = v.(map[string]interface{})
	}
	var interpreter []interface{}
	if v, ok := d.GetOk("interpreter"); ok {
		interpreter = v.([]interface{})
	}
	enableParallelism := d.Get("enable_parallelism").(bool)

	config := Config{
		Environment:          environment,
		SensitiveEnvironment: sensitiveEnvironment,
		Interpreter:          interpreter,
		EnableParallelism:    enableParallelism,
	}
	return config.Client()
}

// This is a global Mutex for use within this plugin.
var shellMutex sync.Mutex
