package shell

import (
	"fmt"
	"log"
	"reflect"
	"runtime"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/rs/xid"
)

func resourceShellScript(sensitive_output bool) *schema.Resource {
	return &schema.Resource{
		Create: resourceShellScriptCreate,
		Delete: resourceShellScriptDelete,
		Read:   resourceShellScriptRead,
		Update: resourceShellScriptUpdate,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		CustomizeDiff: resourceShellScriptCustomizeDiff,
		Schema: map[string]*schema.Schema{
			"lifecycle_commands": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"create": {
							Type:     schema.TypeString,
							Required: true,
						},
						"update": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"read": {
							Type:     schema.TypeString,
							Optional: true,
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
				Sensitive: sensitive_output,
			},
			"dirty": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"read_error": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
				ValidateFunc: func(i interface{}, s string) ([]string, []error) {
					x0 := fmt.Errorf("`read_error` cannot be set from configuration")
					return []string{}, []error{x0}
				},
			},
		},
	}
}

//helpers to unwravel the recursive bits by adding a base condition
func resourceShellScriptCreate(d *schema.ResourceData, meta interface{}) error {
	return create(d, meta, []Action{ActionCreate})
}

func resourceShellScriptRead(d *schema.ResourceData, meta interface{}) error {
	err := read(d, meta, []Action{ActionRead})
	if err != nil {
		_ = d.Set("read_error", err.Error())
	} else {
		_ = d.Set("read_error", "")
	}

	// Error could be caused by bugs in the script.
	// Give user chance to fix the script or continue to recreate the resource
	return nil
}

func resourceShellScriptUpdate(d *schema.ResourceData, meta interface{}) (err error) {
	if d.HasChange("lifecycle_commands.0.read") {
		err = read(d, meta, []Action{ActionUpdate})
		if err != nil {
			_ = restoreOldResourceData(d)
		}
		return
	}

	if d.HasChanges("lifecycle_commands", "interpreter") {
		return
	}

	return update(d, meta, []Action{ActionUpdate})
}

func resourceShellScriptDelete(d *schema.ResourceData, meta interface{}) error {
	return delete(d, meta, []Action{ActionDelete})
}

func resourceShellScriptCustomizeDiff(d *schema.ResourceDiff, i interface{}) (err error) {
	if d.Id() == "" {
		return
	}

	if d.HasChange("lifecycle_commands") || d.HasChange("interpreter") {
		for _, k := range []string{
			"environment", "sensitive_environment", "working_directory"} {

			if d.HasChange(k) {
				return fmt.Errorf("changes to `lifecycle_commands` and/or `interpreter`" +
					" should not be follwed by changes to other arguments")
			}
		}

		if d.HasChange("lifecycle_commands.0.read") {
			_ = d.SetNewComputed("output") //  assume output will change
		}
		return
	}

	if v, _ := d.GetChange("read_error"); v != nil {
		if e, _ := v.(string); e != "" {
			_ = d.ForceNew("read_error") // read error, force recreation
			return
		}
	}

	if _, ok := d.GetOk("lifecycle_commands.0.update"); ok {
		for _, k := range []string{
			"environment", "sensitive_environment", "working_directory"} {

			if d.HasChange(k) {
				_ = d.SetNewComputed("output") //  assume output will change
			}
		}
		return // updateable
	}

	// all the other arguments
	for _, k := range []string{
		"environment", "sensitive_environment", "working_directory", "dirty"} {

		if !d.HasChange(k) {
			continue
		}
		err = d.ForceNew(k)
		if err != nil {
			return
		}
	}
	return
}

func create(d *schema.ResourceData, meta interface{}, stack []Action) error {
	log.Printf("[DEBUG] Creating shell script resource...")
	printStackTrace(stack)

	l := d.Get("lifecycle_commands").([]interface{})
	c := l[0].(map[string]interface{})
	command := c["create"].(string)
	client := meta.(*Client)
	envVariables := getEnvironmentVariables(client, d)
	environment := formatEnvironmentVariables(envVariables)
	sensitiveEnvVariables := getSensitiveEnvironmentVariables(client, d)
	sensitiveEnvironment := formatEnvironmentVariables(sensitiveEnvVariables)
	interpreter := getInterpreter(client, d)
	workingDirectory := d.Get("working_directory").(string)
	enableParallelism := client.config.EnableParallelism
	d.MarkNewResource()

	commandConfig := &CommandConfig{
		Command:              command,
		Environment:          environment,
		SensitiveEnvironment: sensitiveEnvironment,
		WorkingDirectory:     workingDirectory,
		Interpreter:          interpreter,
		Action:               ActionCreate,
		EnableParallelism:    enableParallelism,
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

	//create random uuid for the id only if output has been set
	savedOutput := d.Get("output")
	if savedOutput != nil {
		id := xid.New().String()
		d.SetId(id)
	}

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
	environment := formatEnvironmentVariables(envVariables)
	sensitiveEnvVariables := getSensitiveEnvironmentVariables(client, d)
	sensitiveEnvironment := formatEnvironmentVariables(sensitiveEnvVariables)
	interpreter := getInterpreter(client, d)
	workingDirectory := d.Get("working_directory").(string)
	o, _ := d.GetChange("output")
	previousOutput := expandOutput(o)
	enableParallelism := client.config.EnableParallelism

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
		log.Printf("[DEBUG] State from read operation was nil. Marking resource for deletion.")
		d.SetId("")
		return nil
	}
	log.Printf("[DEBUG] previous output:|%v|", previousOutput)
	log.Printf("[DEBUG] new output:|%v|", output)
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

func restoreOldResourceData(rd *schema.ResourceData, except ...string) (err error) {
	exceptMap := map[string]bool{}
	for _, k := range except {
		exceptMap[k] = true
	}
	for _, k := range []string{
		"lifecycle_commands",
		"triggers",

		"environment",
		"sensitive_environment",
		"interpreter",
		"working_directory",
		"output",

		"dirty",
		"read_error",
	} {

		if exceptMap[k] {
			continue
		}

		o, _ := rd.GetChange(k)
		err = rd.Set(k, o)
		if err != nil {
			return
		}
	}

	return
}

func update(d *schema.ResourceData, meta interface{}, stack []Action) error {
	log.Printf("[DEBUG] Updating shell script resource...")
	d.Set("dirty", false)
	printStackTrace(stack)
	l := d.Get("lifecycle_commands").([]interface{})
	c := l[0].(map[string]interface{})
	command := c["update"].(string)

	client := meta.(*Client)
	envVariables := getEnvironmentVariables(client, d)
	environment := formatEnvironmentVariables(envVariables)
	sensitiveEnvVariables := getSensitiveEnvironmentVariables(client, d)
	sensitiveEnvironment := formatEnvironmentVariables(sensitiveEnvVariables)
	interpreter := getInterpreter(client, d)
	workingDirectory := d.Get("working_directory").(string)
	o, _ := d.GetChange("output")
	previousOutput := expandOutput(o)
	enableParallelism := client.config.EnableParallelism

	commandConfig := &CommandConfig{
		Command:              command,
		Environment:          environment,
		SensitiveEnvironment: sensitiveEnvironment,
		WorkingDirectory:     workingDirectory,
		Interpreter:          interpreter,
		Action:               ActionDelete,
		PreviousOutput:       previousOutput,
		EnableParallelism:    enableParallelism,
	}
	output, err := runCommand(commandConfig)
	if err != nil {
		_ = restoreOldResourceData(d)
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
	if e, _ := d.Get("read_error").(string); e != "" {
		return nil
	}

	log.Printf("[DEBUG] Deleting shell script resource...")
	printStackTrace(stack)
	l := d.Get("lifecycle_commands").([]interface{})
	c := l[0].(map[string]interface{})
	command := c["delete"].(string)

	client := meta.(*Client)
	envVariables := getEnvironmentVariables(client, d)
	environment := formatEnvironmentVariables(envVariables)
	sensitiveEnvVariables := getSensitiveEnvironmentVariables(client, d)
	sensitiveEnvironment := formatEnvironmentVariables(sensitiveEnvVariables)
	interpreter := getInterpreter(client, d)
	workingDirectory := d.Get("working_directory").(string)
	o, _ := d.GetChange("output")
	previousOutput := expandOutput(o)
	enableParallelism := client.config.EnableParallelism

	commandConfig := &CommandConfig{
		Command:              command,
		Environment:          environment,
		SensitiveEnvironment: sensitiveEnvironment,
		WorkingDirectory:     workingDirectory,
		Interpreter:          interpreter,
		Action:               ActionDelete,
		PreviousOutput:       previousOutput,
		EnableParallelism:    enableParallelism,
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

func formatEnvironmentVariables(ev interface{}) []string {
	var envList []string
	envMap := ev.(map[string]interface{})
	if envMap != nil {
		for k, v := range envMap {
			envList = append(envList, k+"="+v.(string))
		}
	}
	return envList
}

func getEnvironmentVariables(client *Client, d *schema.ResourceData) map[string]interface{} {
	variables := make(map[string]interface{})
	resEnv := d.Get("environment").(map[string]interface{})
	for k, v := range client.config.Environment {
		variables[k] = v
	}
	for k, v := range resEnv {
		variables[k] = v
	}

	return variables
}

func getInterpreter(client *Client, d *schema.ResourceData) []string {
	var interpreter []string

	//use provider interpreter if set
	var providerInterpreter []string
	for _, v := range client.config.Interpreter {
		providerInterpreter = append(providerInterpreter, v.(string))
	}
	interpreter = providerInterpreter
	log.Printf("[DEBUG] provider interpreter\n %v", interpreter)
	//override provider supplied interpreter with resource interpreter if set
	var resourceInterpreter []string
	if ir, ok := d.GetOk("interpreter"); ok {
		for _, v := range ir.([]interface{}) {
			resourceInterpreter = append(resourceInterpreter, v.(string))
		}
		interpreter = resourceInterpreter
	}
	log.Printf("[DEBUG] resource interpreter\n %v", interpreter)
	//use default interpreter if none provided
	if interpreter == nil {
		var shell, flag string
		if runtime.GOOS == "windows" {
			shell = "cmd"
			flag = "/C"
		} else {
			shell = "/bin/sh"
			flag = "-c"
		}
		interpreter = []string{shell, flag}
	}
	log.Printf("[DEBUG] interpreter\n %v", interpreter)
	return interpreter
}

func getSensitiveEnvironmentVariables(client *Client, d *schema.ResourceData) map[string]interface{} {
	variables := make(map[string]interface{})
	resEnv := d.Get("sensitive_environment").(map[string]interface{})
	for k, v := range client.config.SensitiveEnvironment {
		variables[k] = v
	}
	for k, v := range resEnv {
		variables[k] = v
	}

	return variables
}
