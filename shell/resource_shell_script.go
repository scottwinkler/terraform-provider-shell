package shell

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
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
		},
	}
}

func resourceShellScriptCreate(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Creating shell script resource...")
	l := d.Get("lifecycle_commands").([]interface{})
	c := l[0].(map[string]interface{})
	command := c["create"].(string)
	vars := d.Get("environment").(map[string]interface{})
	environment := readEnvironmentVariables(vars)
	workingDirectory := d.Get("working_directory").(string)

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
		if err := resourceShellScriptRead(d, meta); err != nil {
			return err
		}
	} else {
		d.Set("output", newState.output)
	}

	//create random uuid for the id
	id := xid.New().String()
	d.SetId(id)
	return nil
}

func resourceShellScriptRead(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Reading shell script resource...")
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
	defer shellMutexKV.Unlock(shellScriptMutexKey)

	state := NewState(environment, output)
	newState, err := runCommand(command, state, environment, workingDirectory)
	if err != nil {
		return err
	}

	if newState == nil {
		return fmt.Errorf("Error: state from read operation cannot be nil")
	}
	d.Set("output", newState.output)
	return nil
}

func resourceShellScriptUpdate(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Updating shell script resource...")
	l := d.Get("lifecycle_commands").([]interface{})
	c := l[0].(map[string]interface{})
	command := c["update"].(string)

	//if update is not set, then treat it simply as a tainted resource - delete then recreate
	if len(command) == 0 {
		resourceShellScriptDelete(d, meta)
		return resourceShellScriptCreate(d, meta)
	}

	//need to get the old environment
	var vars map[string]interface{}
	if d.HasChange("environment") {
		e, _ := d.GetChange("environment")
		vars = e.(map[string]interface{})
	}

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

	//if update doesn't return a new state then must call the read operation
	if newState == nil {
		if err := resourceShellScriptRead(d, meta); err != nil {
			return err
		}
	} else {
		d.Set("output", newState.output)
	}
	return nil
}

func resourceShellScriptDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Deleting shell script resource...")
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
