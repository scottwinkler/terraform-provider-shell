package shell

import (
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/rs/xid"
)

func dataSourceShellScript() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceShellScriptRead,

		Schema: map[string]*schema.Schema{
			"lifecycle_commands": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"read": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
					},
				},
			},
			"environment": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
				Elem:     schema.TypeString,
			},
			"sensitive_environment": {
				Type:      schema.TypeMap,
				Optional:  true,
				Elem:      schema.TypeString,
				Sensitive: true,
			},
			"working_directory": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  ".",
			},
			"output": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     schema.TypeString,
			},
		},
	}
}

func dataSourceShellScriptRead(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Reading shell script data resource...")
	l := d.Get("lifecycle_commands").([]interface{})
	c := l[0].(map[string]interface{})
	value := c["read"]

	command := value.(string)
	environment := packEnvironmentVariables(d.Get("environment"))
	sensitiveEnvironment := packEnvironmentVariables(d.Get("sensitive_environment"))
	workingDirectory := d.Get("working_directory").(string)
	output := make(map[string]string)

	state := NewState(environment, sensitiveEnvironment, output)
	newState, err := runCommand(command, workingDirectory, ReadAction, state )
	if err != nil {
		return err
	}

	if newState == nil {
		log.Printf("[DEBUG] State from read operation was nil. Marking resource for deletion.")
		d.SetId("")
		return nil
	}
	d.Set("output", newState.Output)

	//create random uuid for the id
	id := xid.New().String()
	d.SetId(id)
	return nil
}
