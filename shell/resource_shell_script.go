package shell

import (
	"crypto/rand"
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
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
	l := d.Get("lifecycle_commands").([]interface{})
	c := l[0].(map[string]interface{})
	command := c["create"].(string)
	vars := d.Get("environment").(map[string]interface{})
	environment := readEnvironmentVariables(vars)

	workingDirectory := d.Get("working_directory").(string)

	//obtain exclusive lock
	shellMutexKV.Lock(shellScriptMutexKey)
	defer shellMutexKV.Unlock(shellScriptMutexKey)

	output := make(map[string]string)
	state := NewState(environment, output)
	newState, err := runCommand(command, state, environment, workingDirectory)
	if err != nil {
		return err
	}

	//once creation has finished setup state so update works
	if newState == nil {
		//otherwise set state fix the read()
		if err := resourceShellScriptRead(d, meta); err != nil {
			return err
		}
	} else {
		d.Set("output", newState.output)
	}

	//create random uuid for the id, changes to inputs will prompt update or recreate so it doesn't need to change
	idBytes := make([]byte, 16)
	_, randErr := rand.Read(idBytes)
	if randErr != nil {
		return randErr
	}
	d.SetId(hash(string(idBytes[0:])))
	return nil
}

func resourceShellScriptRead(d *schema.ResourceData, meta interface{}) error {
	l := d.Get("lifecycle_commands").([]interface{})
	c := l[0].(map[string]interface{})
	value := c["read"]

	//if read is not set then do nothing. assume create is setting the state and it never needs to
	//be synchronized with the remote state
	if value == nil {
		return nil
	}

	command := value.(string)
	vars := d.Get("environment").(map[string]interface{})
	environment := readEnvironmentVariables(vars)
	workingDirectory := d.Get("working_directory").(string)
	output := d.Get("output").(map[string]string)

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
	l := d.Get("lifecycle_commands").([]interface{})
	c := l[0].(map[string]interface{})
	value := c["update"]

	//if update is not set, then treat it simply as a tainted resource
	if value == nil {
		resourceShellScriptDelete(d, meta)
		return resourceShellScriptCreate(d, meta)
	}

	command := value.(string)
	vars := d.Get("environment").(map[string]interface{})
	environment := readEnvironmentVariables(vars)
	workingDirectory := d.Get("working_directory").(string)
	output := d.Get("output").(map[string]string)

	//obtain exclusive lock
	shellMutexKV.Lock(shellScriptMutexKey)
	defer shellMutexKV.Unlock(shellScriptMutexKey)

	state := NewState(environment, output)
	newState, err := runCommand(command, state, environment, workingDirectory)
	if err != nil {
		return err
	}

	//once creation has finished setup state so update works
	if newState == nil {
		//otherwise set state fix the read()
		if err := resourceShellScriptRead(d, meta); err != nil {
			return err
		}
	} else {
		d.Set("output", newState.output)
	}
	return nil
}

func resourceShellScriptDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Deleting shell script resource")
	l := d.Get("lifecycle_commands").([]interface{})
	c := l[0].(map[string]interface{})
	command := c["delete"].(string)
	vars := d.Get("environment").(map[string]interface{})
	environment := readEnvironmentVariables(vars)
	workingDirectory := d.Get("working_directory").(string)
	output := d.Get("output").(map[string]string)

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
