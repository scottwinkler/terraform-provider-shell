package shell

import (
	"log"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceShellScript() *schema.Resource {
	return &schema.Resource{
		Create: resourceShellScriptCreate,
		Delete: resourceShellScriptDelete,
		Read:   resourceShellScriptRead,
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
						"read": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							Default:  "IN=$(cat)\necho $IN",
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
			"stderr": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"stdout": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"output": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     schema.TypeString,
			},
		},
	}
}

func resourceShellScriptDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Deleting shell script resource")
	l := d.Get("lifecycle_commands").([]interface{})
	c := l[0].(map[string]interface{})
	command := c["delete"].(string)
	vars := d.Get("environment").(map[string]interface{})
	environment := readEnvironmentVariables(vars)
	workingDirectory := d.Get("working_directory").(string)
	input := d.Get("stdout").(string)
	//obtain exclusive lock
	shellMutexKV.Lock(shellScriptMutexKey)
	defer shellMutexKV.Unlock(shellScriptMutexKey)

	_, _, err := runCommand(command, input, environment, workingDirectory)
	if err != nil {
		return err
	}

	d.SetId("")
	return nil
}
func resourceShellScriptCreate(d *schema.ResourceData, meta interface{}) error {
	l := d.Get("lifecycle_commands").([]interface{})
	c := l[0].(map[string]interface{})
	command := c["create"].(string)
	vars := d.Get("environment").(map[string]interface{})
	environment := readEnvironmentVariables(vars)
	workingDirectory := d.Get("working_directory").(string)
	var input string
	//obtain exclusive lock
	shellMutexKV.Lock(shellScriptMutexKey)
	defer shellMutexKV.Unlock(shellScriptMutexKey)

	stdout, stderr, err := runCommand(command, input, environment, workingDirectory)
	if err != nil {
		return err
	}
	output, err := parseJSON(stdout)
	if err != nil {
		log.Printf("[DEBUG] error parsing sdout into json: %v", err)
		output = make(map[string]string)
		d.Set("output", output)
	} else {
		d.Set("output", output)
	}

	d.Set("stdout", stdout)
	d.Set("stderr", stderr)
	d.SetId(hash(command))
	return nil
}

func resourceShellScriptRead(d *schema.ResourceData, meta interface{}) error {
	l := d.Get("lifecycle_commands").([]interface{})
	c := l[0].(map[string]interface{})
	command := c["read"].(string)
	vars := d.Get("environment").(map[string]interface{})
	environment := readEnvironmentVariables(vars)
	workingDirectory := d.Get("working_directory").(string)
	input := d.Get("stdout").(string)

	//obtain exclusive lock
	shellMutexKV.Lock(shellScriptMutexKey)
	defer shellMutexKV.Unlock(shellScriptMutexKey)

	stdout, stderr, err := runCommand(command, input, environment, workingDirectory)
	if err != nil {
		return err
	}
	output, err := parseJSON(stdout)
	if err != nil {
		log.Printf("[DEBUG] error parsing sdout into json: %v", err)
		var output map[string]string
		d.Set("output", output)
	} else {
		d.Set("output", output)
	}

	d.Set("stdout", stdout)
	d.Set("stderr", stderr)
	return nil
}
