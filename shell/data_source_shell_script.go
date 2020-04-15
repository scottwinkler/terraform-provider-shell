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

	//we don't care about previous output for data sources
	previousOutput := make(map[string]string)

	commandConfig := &CommandConfig{
		Command:              command,
		Environment:          environment,
		SensitiveEnvironment: sensitiveEnvironment,
		WorkingDirectory:     workingDirectory,
		Action:               ActionRead,
		PreviousOutput:       previousOutput,
	}
	output, err := runCommand(commandConfig)
	if err != nil {
		return err
	}

	if output == nil {
		log.Printf("[DEBUG] Output from read operation was nil. Marking resource for deletion.")
		d.SetId("")
		return nil
	}
	d.Set("output", output)

	//create random uuid for the id
	id := xid.New().String()
	d.SetId(id)
	return nil
}
