package shell

import (
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/rs/xid"
)

func dataSourceShellScript() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceShellScriptRead,

		Schema: map[string]*schema.Schema{
			"lifecycle_commands": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"read": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"environment": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     schema.TypeString,
			},
			"sensitive_environment": {
				Type:      schema.TypeMap,
				Optional:  true,
				Elem:      schema.TypeString,
				Sensitive: true,
			},
			"interpreter": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"working_directory": {
				Type:     schema.TypeString,
				Optional: true,
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
	client := meta.(*Client)
	envVariables := getEnvironmentVariables(client, d)
	environment := formatEnvironmentVariables(envVariables)
	sensitiveEnvVariables := getSensitiveEnvironmentVariables(client, d)
	sensitiveEnvironment := formatEnvironmentVariables(sensitiveEnvVariables)
	interpreter := getInterpreter(client, d)
	workingDirectory := d.Get("working_directory").(string)
	enableParallelism := client.config.EnableParallelism

	//we don't care about previous output for data sources
	previousOutput := make(map[string]string)

	commandConfig := &CommandConfig{
		Command:              command,
		Environment:          environment,
		SensitiveEnvironment: sensitiveEnvironment,
		WorkingDirectory:     workingDirectory,
		Interpreter:          interpreter,
		Action:               ActionRead,
		PreviousOutput:       previousOutput,
		EnableParallelism:    enableParallelism,
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
