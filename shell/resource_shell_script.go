package shell

import (
	"log"
	"reflect"

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
							ForceNew: true,
						},
						"read": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"delete": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
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
				ForceNew: true,
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
	return create(d, meta, []string{"create"})
}

func resourceShellScriptRead(d *schema.ResourceData, meta interface{}) error {
	return read(d, meta, []string{"read"})
}

func resourceShellScriptUpdate(d *schema.ResourceData, meta interface{}) error {
	return update(d, meta, []string{"update"})
}

func resourceShellScriptDelete(d *schema.ResourceData, meta interface{}) error {
	return delete(d, meta, []string{"delete"})
}

func create(d *schema.ResourceData, meta interface{}, stack []string) error {
	log.Printf("[DEBUG] Creating shell script resource...")
	printStackTrace(stack)
	l := d.Get("lifecycle_commands").([]interface{})
	c := l[0].(map[string]interface{})
	command := c["create"].(string)
	environment := packEnvironmentVariables(d.Get("environment"))
	sensitiveEnvironment := packEnvironmentVariables(d.Get("sensitive_environment"))
	workingDirectory := d.Get("working_directory").(string)
	d.MarkNewResource()
	output := make(map[string]string)
	state := NewState(environment, sensitiveEnvironment, output)
	newState, err := runCommand(command, state, workingDirectory)
	if err != nil {
		return err
	}

	//if create doesn't return a new state then must call the read operation
	if newState == nil {
		stack = append(stack, "read")
		if err := read(d, meta, stack); err != nil {
			return err
		}
	} else {
		d.Set("output", newState.Output)
	}

	//create random uuid for the id
	id := xid.New().String()
	d.SetId(id)
	return nil
}

func read(d *schema.ResourceData, meta interface{}, stack []string) error {
	log.Printf("[DEBUG] Reading shell script resource...")
	printStackTrace(stack)
	l := d.Get("lifecycle_commands").([]interface{})
	c := l[0].(map[string]interface{})
	command := c["read"].(string)

	//if read is not set then do nothing. assume something either create or update is setting the state
	if len(command) == 0 {
		return nil
	}

	environment := packEnvironmentVariables(d.Get("environment"))
	sensitiveEnvironment := packEnvironmentVariables(d.Get("sensitive_environment"))
	workingDirectory := d.Get("working_directory").(string)
	output := expandOutput(d.Get("output"))

	state := NewState(environment, sensitiveEnvironment, output)
	newState, err := runCommand(command, state, workingDirectory)
	if err != nil {
		return err
	}

	if newState == nil {
		log.Printf("[DEBUG] State from read operation was nil. Marking resource for deletion.")
		d.SetId("")
		return nil
	}
	log.Printf("[DEBUG] output:|%v|", output)
	log.Printf("[DEBUG] new output:|%v|", newState.Output)
	isStateEqual := reflect.DeepEqual(output, newState.Output)
	isNewResource := d.IsNewResource()
	isUpdatedResource := stack[0] == "update"
	if !isStateEqual && !isNewResource && !isUpdatedResource {
		log.Printf("[DEBUG] Previous state not equal to new state. Marking resource as dirty to trigger update.")
		d.Set("dirty", true)
		return nil
	}

	d.Set("output", newState.Output)

	return nil
}

func update(d *schema.ResourceData, meta interface{}, stack []string) error {
	log.Printf("[DEBUG] Updating shell script resource...")
	d.Set("dirty", false)
	printStackTrace(stack)
	l := d.Get("lifecycle_commands").([]interface{})
	c := l[0].(map[string]interface{})
	command := c["update"].(string)

	//if update is not set, then treat it simply as a tainted resource - delete then recreate
	if len(command) == 0 {
		stack = append(stack, "delete")
		delete(d, meta, stack)
		stack = append(stack, "create")
		return create(d, meta, stack)
	}

	environment := packEnvironmentVariables(d.Get("environment"))
	sensitiveEnvironment := packEnvironmentVariables(d.Get("sensitive_environment"))
	workingDirectory := d.Get("working_directory").(string)
	output := expandOutput(d.Get("output"))

	state := NewState(environment, sensitiveEnvironment, output)
	newState, err := runCommand(command, state, workingDirectory)
	if err != nil {
		return err
	}

	//if update doesn't return a new state then must call the read operation
	if newState == nil {
		stack = append(stack, "read")
		if err := read(d, meta, stack); err != nil {
			return err
		}
	} else {
		d.Set("output", newState.Output)
	}
	return nil
}

func delete(d *schema.ResourceData, meta interface{}, stack []string) error {
	log.Printf("[DEBUG] Deleting shell script resource...")
	printStackTrace(stack)
	l := d.Get("lifecycle_commands").([]interface{})
	c := l[0].(map[string]interface{})
	command := c["delete"].(string)
	environment := packEnvironmentVariables(d.Get("environment"))
	sensitiveEnvironment := packEnvironmentVariables(d.Get("sensitive_environment"))
	workingDirectory := d.Get("working_directory").(string)
	output := expandOutput(d.Get("output"))

	state := NewState(environment, sensitiveEnvironment, output)
	_, err := runCommand(command, state, workingDirectory)
	if err != nil {
		return err
	}
	d.SetId("")
	return nil
}

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
