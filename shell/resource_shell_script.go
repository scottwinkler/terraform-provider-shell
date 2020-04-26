package shell

import (
	"log"
	"reflect"

	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/rs/xid"
)

func resourceShellScript() *schema.Resource {
	return &schema.Resource{
		Create: resourceShellScriptCreate,
		Delete: resourceShellScriptDelete,
		Read:   resourceShellScriptRead,
		Update: resourceShellScriptUpdate,
		Schema: map[string]*schema.Schema{
			"lifecycle_commands": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"create": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"update": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"read": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"delete": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"triggers": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
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
			"dirty": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
	}
}

//helpers to unwravel the recursive bits by adding a base condition
func resourceShellScriptCreate(d *schema.ResourceData, meta interface{}) error {
	return create(d, meta, []Action{ActionCreate})
}

func resourceShellScriptRead(d *schema.ResourceData, meta interface{}) error {
	return read(d, meta, []Action{ActionRead})
}

func resourceShellScriptUpdate(d *schema.ResourceData, meta interface{}) error {
	return update(d, meta, []Action{ActionUpdate})
}

func resourceShellScriptDelete(d *schema.ResourceData, meta interface{}) error {
	return delete(d, meta, []Action{ActionDelete})
}

func create(d *schema.ResourceData, meta interface{}, stack []Action) error {
	log.Printf("[DEBUG] Creating shell script resource...")
	printStackTrace(stack)

	l := d.Get("lifecycle_commands").([]interface{})
	c := l[0].(map[string]interface{})
	command := c["create"].(string)
	client := meta.(*Client)
	envVariables := getEnvironmentVariables(client, d)
	environment := flattenEnvironmentVariables(envVariables)
	sensitiveEnvironment := flattenEnvironmentVariables(d.Get("sensitive_environment"))

	workingDirectory := d.Get("working_directory").(string)
	d.MarkNewResource()
	commandConfig := &CommandConfig{
		Command:              command,
		Environment:          environment,
		SensitiveEnvironment: sensitiveEnvironment,
		WorkingDirectory:     workingDirectory,
		Action:               ActionCreate,
	}
	output, err := runCommand(commandConfig)
	if err != nil {
		return err
	}

	//if create doesn't return output then assume we should call the read operation
	if output == nil {
		stack = append(stack, ActionRead)
		if err := read(d, meta, stack); err != nil {
			return err
		}
	} else {
		d.Set("output", output)
	}

	//create random uuid for the id
	id := xid.New().String()
	d.SetId(id)
	return nil
}

func read(d *schema.ResourceData, meta interface{}, stack []Action) error {
	log.Printf("[DEBUG] Reading shell script resource...")
	printStackTrace(stack)
	l := d.Get("lifecycle_commands").([]interface{})
	c := l[0].(map[string]interface{})
	command := c["read"].(string)

	//if read is not set then do nothing. assume something either create or update is setting the state
	if len(command) == 0 {
		return nil
	}

	client := meta.(*Client)
	envVariables := getEnvironmentVariables(client, d)
	environment := flattenEnvironmentVariables(envVariables)
	sensitiveEnvironment := flattenEnvironmentVariables(d.Get("sensitive_environment"))
	workingDirectory := d.Get("working_directory").(string)
	previousOutput := expandOutput(d.Get("output"))

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
		log.Printf("[DEBUG] State from read operation was nil. Marking resource for deletion.")
		d.SetId("")
		return nil
	}
	log.Printf("[DEBUG] previous output:|%v|", output)
	log.Printf("[DEBUG] new output:|%v|", previousOutput)
	isStateEqual := reflect.DeepEqual(output, previousOutput)
	isNewResource := d.IsNewResource()
	isUpdatedResource := stack[0] == ActionUpdate
	if !isStateEqual && !isNewResource && !isUpdatedResource {
		log.Printf("[DEBUG] Previous state not equal to new state. Marking resource as dirty to trigger update.")
		d.Set("dirty", true)
		return nil
	}

	d.Set("output", output)

	return nil
}

func update(d *schema.ResourceData, meta interface{}, stack []Action) error {
	log.Printf("[DEBUG] Updating shell script resource...")
	d.Set("dirty", false)
	printStackTrace(stack)
	l := d.Get("lifecycle_commands").([]interface{})
	c := l[0].(map[string]interface{})
	command := c["update"].(string)

	//if update is not set, then treat it simply as a tainted resource - delete then recreate
	if len(command) == 0 {
		stack = append(stack, ActionDelete)
		delete(d, meta, stack)
		stack = append(stack, ActionCreate)
		return create(d, meta, stack)
	}

	client := meta.(*Client)
	envVariables := getEnvironmentVariables(client, d)
	environment := flattenEnvironmentVariables(envVariables)
	sensitiveEnvironment := flattenEnvironmentVariables(d.Get("sensitive_environment"))
	workingDirectory := d.Get("working_directory").(string)
	previousOutput := expandOutput(d.Get("output"))

	commandConfig := &CommandConfig{
		Command:              command,
		Environment:          environment,
		SensitiveEnvironment: sensitiveEnvironment,
		WorkingDirectory:     workingDirectory,
		Action:               ActionDelete,
		PreviousOutput:       previousOutput,
	}
	output, err := runCommand(commandConfig)
	if err != nil {
		return err
	}

	//if update doesn't return output then must call the read operation
	if output == nil {
		stack = append(stack, ActionRead)
		if err := read(d, meta, stack); err != nil {
			return err
		}
	} else {
		d.Set("output", output)
	}
	return nil
}

func delete(d *schema.ResourceData, meta interface{}, stack []Action) error {
	log.Printf("[DEBUG] Deleting shell script resource...")
	printStackTrace(stack)
	l := d.Get("lifecycle_commands").([]interface{})
	c := l[0].(map[string]interface{})
	command := c["delete"].(string)

	client := meta.(*Client)
	envVariables := getEnvironmentVariables(client, d)
	environment := flattenEnvironmentVariables(envVariables)
	sensitiveEnvironment := flattenEnvironmentVariables(d.Get("sensitive_environment"))
	workingDirectory := d.Get("working_directory").(string)
	previousOutput := expandOutput(d.Get("output"))

	commandConfig := &CommandConfig{
		Command:              command,
		Environment:          environment,
		SensitiveEnvironment: sensitiveEnvironment,
		WorkingDirectory:     workingDirectory,
		Action:               ActionDelete,
		PreviousOutput:       previousOutput,
	}
	_, err := runCommand(commandConfig)
	if err != nil {
		return err
	}
	d.SetId("")
	return nil
}

//Action is an enum for CRUD operations
type Action string

const (
	//ActionCreate ...
	ActionCreate Action = "create"
	//ActionRead ...
	ActionRead Action = "read"
	//ActionUpdate ...
	ActionUpdate Action = "update"
	//ActionDelete ...
	ActionDelete Action = "delete"
)

func expandOutput(o interface{}) map[string]string {
	om := o.(map[string]interface{})
	output := make(map[string]string)
	if om != nil {
		for k, v := range om {
			output[k] = v.(string)
		}
	}
	return output
}

func flattenEnvironmentVariables(ev interface{}) []string {
	var envList []string
	envMap := ev.(map[string]interface{})
	if envMap != nil {
		for k, v := range envMap {
			envList = append(envList, k+"="+v.(string))
		}
	}
	return envList
}

func expandEnvironmentVariables(envList []string) map[string]string {
	envMap := make(map[string]string)
	for _, v := range envList {
		parts := strings.Split(v, "=")
		key := parts[0]
		value := parts[1]
		envMap[key] = value
	}
	return envMap
}
