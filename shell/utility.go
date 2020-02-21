package shell

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime"

	"github.com/armon/circbuf"
)

// State is a wrapper around both the input and output attributes that are relavent for updates
type State struct {
	Environment []string
	Output      map[string]string
}

// NewState is the constructor for State
func NewState(environment []string, output map[string]string) *State {
	return &State{Environment: environment, Output: output}
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

func readInterpreterVariables(ev map[string]interface{}) [2]string {
	var variables [2]string
	if ev != nil {
		for k := range ev {
			if k == "flag" {
				variables[0] = ev[k].(string)
			} else if k == "shell" {
				variables[1] = ev[k].(string)
			}
		}

	}
	return variables
}

func printStackTrace(stack []string) {
	log.Printf("-------------------------")
	log.Printf("[DEBUG] Current stack:")
	for _, v := range stack {
		log.Printf("[DEBUG] -- %s", v)
	}
	log.Printf("-------------------------")
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

func runCommand(command string, state *State, environment []string, workingDirectory string, interpreter [2]string) (*State, error) {
	const maxBufSize = 8 * 1024
	// Execute the command using a shell
	var shell, flag string
	if interpreter[0] != "" {
		shell = interpreter[1]
		flag = interpreter[0]
	} else {
		if runtime.GOOS == "windows" {
			shell = "cmd"
			flag = "/C"
		} else {
			shell = "/bin/sh"
			flag = "-c"
		}
	}

	// Setup the command
	command = fmt.Sprintf("cd %s && %s", workingDirectory, command)
	cmd := exec.Command(shell, flag, command)
	input, _ := json.Marshal(state.Output)
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

	log.Printf("[DEBUG] shell script command old state: \"%v\"", state)

	// Output what we're about to run
	log.Printf("[DEBUG] shell script going to execute: %s %s \"%s\"", shell, flag, command)

	// Run the command to completion
	err = cmd.Run()
	pw.Close()
	log.Printf("[DEBUG] Command execution completed. Reading from output pipe: >&3")

	//read back diff output from pipe
	buffer := new(bytes.Buffer)
	for {
		tmpdata := make([]byte, maxBufSize)
		bytecount, _ := pr.Read(tmpdata)
		if bytecount == 0 {
			break
		}
		buffer.Write(tmpdata)
	}
	log.Printf("[DEBUG] shell script command stdout: \"%s\"", stdout.String())
	log.Printf("[DEBUG] shell script command stderr: \"%s\"", stderr.String())
	log.Printf("[DEBUG] shell script command output: \"%s\"", buffer.String())

	if err != nil {
		return nil, fmt.Errorf("Error running command: '%v'", err)
	}

	output, err := parseJSON(buffer.Bytes())
	if err != nil {
		log.Printf("[DEBUG] Unable to unmarshall data to map[string]string: '%v'", err)
		return nil, nil
	}
	newState := NewState(environment, output)
	log.Printf("[DEBUG] shell script command new state: \"%v\"", newState)
	return newState, nil
}
