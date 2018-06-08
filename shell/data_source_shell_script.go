package shell

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime"

	"github.com/armon/circbuf"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceShellScript() *schema.Resource {
	return &schema.Resource{
		Read: shellScriptRead,

		Schema: map[string]*schema.Schema{
			"read": {
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

func shellScriptRead(d *schema.ResourceData, meta interface{}) error {
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
	d.SetId(hash(command))
	return nil
}

func parseJSON(s string) (map[string]string, error) {
	var f map[string]interface{}
	err := json.Unmarshal([]byte(s), &f)
	output := make(map[string]string)
	for k, v := range f {
		output[k] = v.(string)
	}
	return output, err
}

func readEnvironmentVariables(ev map[string]interface{}) []string {
	var variables []string
	variables = append(os.Environ())
	if ev != nil {
		for k, v := range ev {
			variables = append(variables, k+"="+v.(string))
		}
	}
	return variables
}

func runCommand(command string, environment []string, workingDirectory string) (string, string, error) {
	const maxBufSize = 8 * 1024
	// Execute the command using a shell
	var shell, flag string
	if runtime.GOOS == "windows" {
		shell = "cmd"
		flag = "/C"
	} else {
		shell = "/bin/sh"
		flag = "-c"
	}

	// Setup the command
	command = fmt.Sprintf("cd %s && %s", workingDirectory, command)
	cmd := exec.Command(shell, flag, command)
	cmd.Env = environment
	stdout, _ := circbuf.NewBuffer(maxBufSize)
	stderr, _ := circbuf.NewBuffer(maxBufSize)
	cmd.Stderr = io.Writer(stderr)
	cmd.Stdout = io.Writer(stdout)

	// Output what we're about to run
	log.Printf("[DEBUG] shell script going to execute: %s %s \"%s\"", shell, flag, command)

	// Run the command to completion
	err := cmd.Run()

	if err != nil {
		return "", "", fmt.Errorf("Error running command '%s': '%v'", command, err)
	}

	log.Printf("[DEBUG] shell script command stdout was: \"%s\"", stdout.String())
	log.Printf("[DEBUG] shell script command stderr was: \"%s\"", stderr.String())
	return stdout.String(), stderr.String(), nil
}
func hash(s string) string {
	sha := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sha[:])
}
