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
			//possibly make this a map called commands?
			"create": {
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
	command := d.Get("delete").(string)
	vars := d.Get("environment").(map[string]interface{})
	environment := readEnvironmentVariables(vars)
	workingDirectory := d.Get("working_directory").(string)

	//obtain exclusive lock
	shellMutexKV.Lock(shellScriptMutexKey)
	defer shellMutexKV.Unlock(shellScriptMutexKey)

	_, _, err := runCommand(command, environment, workingDirectory)
	if err != nil {
		return err
	}

	d.SetId("")
	return nil
}
func resourceShellScriptCreate(d *schema.ResourceData, meta interface{}) error {
	command := d.Get("create").(string)
	vars := d.Get("environment").(map[string]interface{})
	environment := readEnvironmentVariables(vars)
	workingDirectory := d.Get("working_directory").(string)

	//obtain exclusive lock
	shellMutexKV.Lock(shellScriptMutexKey)
	defer shellMutexKV.Unlock(shellScriptMutexKey)

	stdout, stderr, err := runCommand(command, environment, workingDirectory)
	if err != nil {
		return err
	}
	output, err := parseJSON(stdout)
	if err != nil {
		log.Printf("[DEBUG] error parsing sdout into json: %v", err)
	} else {
		d.Set("output", output)
	}

	d.Set("stdout", stdout)
	d.Set("stderr", stderr)
	d.SetId(hash(command))
	return nil
}

func resourceShellScriptRead(d *schema.ResourceData, meta interface{}) error {
	command := d.Get("read").(string)
	vars := d.Get("environment").(map[string]interface{})
	environment := readEnvironmentVariables(vars)
	workingDirectory := d.Get("working_directory").(string)

	//obtain exclusive lock
	shellMutexKV.Lock(shellScriptMutexKey)
	defer shellMutexKV.Unlock(shellScriptMutexKey)

	stdout, stderr, err := runCommand(command, environment, workingDirectory)
	if err != nil {
		return err
	}
	output, err := parseJSON(stdout)
	if err != nil {
		log.Printf("[DEBUG] error parsing sdout into json: %v", err)
	} else {
		d.Set("output", output)
	}

	d.Set("stdout", stdout)
	d.Set("stderr", stderr)
	return nil
}
