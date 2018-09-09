package shell

import (
	"log"
	"github.com/hashicorp/terraform/helper/schema"
	"crypto/rand"
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
							Required: true,
							ForceNew: true,
						},
						"read": {
							Type:     schema.TypeString,
							Required: true,
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
				ForceNew: true,
				Elem:     schema.TypeString,
			},
			"working_directory": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  ".",
			},
			"state": {
				Type:     schema.TypeString,
				Optional: true,
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
	
	extraout, stdout, _, err := runCommand(command, "", environment, workingDirectory)
	if err != nil {
		return err
	}
	log.Printf(stdout)

	//try and parse extraout as JSON to expose, could just be a string
	output, err := parseJSON(extraout)
	if err != nil {
		log.Printf("[DEBUG] error parsing sdout into json: %v", err)
		output = make(map[string]string)
		d.Set("output", output)
	} else {
		d.Set("output", output)
	}

	//create random uuid for the id, changes to inputs will prompt update or recreate so it doesn't need to change
	idBytes := make([]byte, 16)
    _, randErr := rand.Read(idBytes)
    if randErr != nil {
        return randErr
    }
    d.SetId(hash(string(idBytes[0:])))

	//once creation has finished setup state so update works
	if len(extraout) > 0 {
		//if we have some content on extraout (used for diff) use that
		d.Set("state", extraout)
		shellMutexKV.Unlock(shellScriptMutexKey)
		return nil

	} else {
		shellMutexKV.Unlock(shellScriptMutexKey)
		//otherwise set state fix the read()
		return resourceShellScriptRead(d, meta)
	}
	
}

func resourceShellScriptRead(d *schema.ResourceData, meta interface{}) error {
	l := d.Get("lifecycle_commands").([]interface{})
	c := l[0].(map[string]interface{})
	command := c["read"].(string)
	vars := d.Get("environment").(map[string]interface{})
	environment := readEnvironmentVariables(vars)
	workingDirectory := d.Get("working_directory").(string)
	input := d.Get("state").(string)

	//obtain exclusive lock
	shellMutexKV.Lock(shellScriptMutexKey)
	defer shellMutexKV.Unlock(shellScriptMutexKey)

	extraout, _, _, err := runCommand(command, "", environment, workingDirectory)
	if err != nil {
		return err
	}
	output, err := parseJSON(extraout)
	if err != nil {
		log.Printf("[DEBUG] error parsing sdout into json: %v", err)
		var output map[string]string
		d.Set("output", output)
	} else {
		d.Set("output", output)
	}

	if len(extraout) == 0 {
		d.SetId("")
	} else {
		log.Printf("Have state: " + input)
		log.Printf("Setting as state: " + extraout)
		d.Set("state", extraout)
	}
	return nil
}

func resourceShellScriptUpdate(d *schema.ResourceData, meta interface{}) error {
	l := d.Get("lifecycle_commands").([]interface{})
	c := l[0].(map[string]interface{})
	command := c["update"].(string)
	vars := d.Get("environment").(map[string]interface{})
	environment := readEnvironmentVariables(vars)
	workingDirectory := d.Get("working_directory").(string)

	_, _, _, err := runCommand(command, "", environment, workingDirectory)
	if err != nil {
		return err
	}

    return resourceShellScriptRead(d, meta)
}

func resourceShellScriptDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Deleting shell script resource")
	l := d.Get("lifecycle_commands").([]interface{})
	c := l[0].(map[string]interface{})
	command := c["delete"].(string)
	vars := d.Get("environment").(map[string]interface{})
	environment := readEnvironmentVariables(vars)
	workingDirectory := d.Get("working_directory").(string)
	//obtain exclusive lock
	shellMutexKV.Lock(shellScriptMutexKey)
	defer shellMutexKV.Unlock(shellScriptMutexKey)

	_, _, _, err := runCommand(command, "", environment, workingDirectory)
	if err != nil {
		return err
	}

	d.SetId("")
	return nil
}