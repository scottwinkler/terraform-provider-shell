package shell

import (
	"bytes"
	"crypto/rand"
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
		Read: dataSourceShellScriptRead,

		Schema: map[string]*schema.Schema{
			"lifecycle_commands": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"read": {
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
			"output": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     schema.TypeString,
			},
		},
	}
}

func dataSourceShellScriptRead(d *schema.ResourceData, meta interface{}) error {
	l := d.Get("lifecycle_commands").([]interface{})
	c := l[0].(map[string]interface{})
	value := c["read"]

	command := value.(string)
	vars := d.Get("environment").(map[string]interface{})
	environment := readEnvironmentVariables(vars)
	workingDirectory := d.Get("working_directory").(string)
	output := make(map[string]string)

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

	//create random uuid for the id, changes to inputs will prompt update or recreate so it doesn't need to change
	idBytes := make([]byte, 16)
	_, randErr := rand.Read(idBytes)
	if randErr != nil {
		return randErr
	}
	d.SetId(hash(string(idBytes[0:])))
	return nil
}

func readEnvironmentVariables(ev map[string]interface{}) []string {
	var variables []string
	if ev != nil {
		for k, v := range ev {
			variables = append(variables, k+"="+v.(string))
		}
	}
	return variables
}

func hash(s string) string {
	sha := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sha[:])
}

func parseJSON(b []byte) (map[string]string, error) {
	tb := bytes.Trim(b, "\x00")
	s := string(tb)
	var f map[string]interface{}
	err := json.Unmarshal([]byte(s), &f)
	output := make(map[string]string)
	for k, v := range f {
		output[k] = v.(string)
	}
	return output, err
}

func runCommand(command string, state *State, environment []string, workingDirectory string) (*State, error) {
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
	input, _ := json.Marshal(state)
	stdin := bytes.NewReader(input)
	cmd.Stdin = stdin
	environment = append(environment, os.Environ()...)
	cmd.Env = environment
	stdout, _ := circbuf.NewBuffer(maxBufSize)
	stderr, _ := circbuf.NewBuffer(maxBufSize)
	cmd.Stderr = io.Writer(stderr)
	cmd.Stdout = io.Writer(stdout)
	pr, pw, err := os.Pipe()
	cmd.ExtraFiles = []*os.File{pw}

	// Output what we're about to run
	log.Printf("[DEBUG] shell script going to execute: %s %s \"%s\"", shell, flag, command)

	// Run the command to completion
	err = cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("Error running command '%s': '%v'", command, err)
	}

	log.Printf("[DEBUG] Command execution completed. Reading from output pipe: >&3")
	//read back diff output from pipe
	data := make([]byte, maxBufSize)
	pr.Read(data)
	pw.Close()
	pr.Close()
	log.Printf("[DEBUG] data raw text from pipe: %s", string(data))
	output, err := parseJSON(data)
	if err != nil {
		return nil, fmt.Errorf("Error unmarshalling data to map[string]string: '%v'", err)
	}
	newState := NewState(environment, output)

	log.Printf("[DEBUG] shell script command stdout: \"%s\"", stdout.String())
	log.Printf("[DEBUG] shell script command stderr: \"%s\"", stderr.String())
	log.Printf("[DEBUG] shell script command state: \"%v\"", newState)
	return newState, nil
}

// State is a wrapper around both the input and output attributes that are relavent for updates
type State struct {
	environment []string          `json:"environment"`
	output      map[string]string `json:"output"`
}

// NewState is the constructor for State
func NewState(environment []string, output map[string]string) *State {
	return &State{environment: environment, output: output}
}
