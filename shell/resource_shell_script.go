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
	vars := d.Get("environment").(map[string]interface{})
	environment := readEnvironmentVariables(vars)
	workingDirectory := d.Get("working_directory").(string)
	d.MarkNewResource()
	//obtain exclusive lock
	shellMutexKV.Lock(shellScriptMutexKey)

	output := make(map[string]string)
	state := NewState(environment, output)
	newState, err := runCommand(command, state, environment, workingDirectory)
	if err != nil {
		return err
	}
	shellMutexKV.Unlock(shellScriptMutexKey)

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

	vars := d.Get("environment").(map[string]interface{})
	environment := readEnvironmentVariables(vars)
	workingDirectory := d.Get("working_directory").(string)
	o := d.Get("output").(map[string]interface{})
	output := make(map[string]string)
	for k, v := range o {
		output[k] = v.(string)
	}

	//obtain exclusive lock
	shellMutexKV.Lock(shellScriptMutexKey)

	state := NewState(environment, output)
	newState, err := runCommand(command, state, environment, workingDirectory)
	if err != nil {
		return err
	}

	shellMutexKV.Unlock(shellScriptMutexKey)
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

	//need to get the old environment if it exists
	oldVars := make(map[string]interface{})
	if d.HasChange("environment") {
		e, _ := d.GetChange("environment")
		oldVars = e.(map[string]interface{})
	}
	oldEnvironment := readEnvironmentVariables(oldVars)
	vars := d.Get("environment").(map[string]interface{})
	environment := readEnvironmentVariables(vars)

	workingDirectory := d.Get("working_directory").(string)
	o := d.Get("output").(map[string]interface{})
	output := make(map[string]string)
	for k, v := range o {
		output[k] = v.(string)
	}

	//obtain exclusive lock
	shellMutexKV.Lock(shellScriptMutexKey)

	state := NewState(oldEnvironment, output)
	newState, err := runCommand(command, state, environment, workingDirectory)
	if err != nil {
		return err
	}

	shellMutexKV.Unlock(shellScriptMutexKey)

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
	vars := d.Get("environment").(map[string]interface{})
	environment := readEnvironmentVariables(vars)
	workingDirectory := d.Get("working_directory").(string)
	o := d.Get("output").(map[string]interface{})
	output := make(map[string]string)
	for k, v := range o {
		output[k] = v.(string)
	}

	//obtain exclusive lock
	shellMutexKV.Lock(shellScriptMutexKey)
	defer shellMutexKV.Unlock(shellScriptMutexKey)

	state := NewState(environment, output)
	_, err := runCommand(command, state, environment, workingDirectory)
	if err != nil {
		return err
	}

	d.SetId("")
	return nil
}
