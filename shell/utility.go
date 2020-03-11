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
	"strings"

	"regexp"

	"github.com/mitchellh/go-linereader"
)

const maxBufSize = 8 * 1024

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

func printStackTrace(stack []string) {
	log.Printf("-------------------------")
	log.Printf("[DEBUG] Current stack:")
	for _, v := range stack {
		log.Printf("[DEBUG] -- %s", v)
	}
	log.Printf("-------------------------")
}

func runCommand(command string, state *State, environment []string, workingDirectory string) (*State, error) {
	shellMutexKV.Lock(shellScriptMutexKey)
	defer shellMutexKV.Unlock(shellScriptMutexKey)

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
	cmd := exec.Command(shell, flag, command)
	input, _ := json.Marshal(state.Output)
	stdin := bytes.NewReader(input)
	cmd.Stdin = stdin
	environment = append(os.Environ(), environment...)
	cmd.Env = environment
	prStdout, pwStdout, err := os.Pipe()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize pipe for stdout: %s", err)
	}
	cmd.Stdout = pwStdout
	prStderr, pwStderr, err := os.Pipe()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize pipe for stderr: %s", err)
	}
	cmd.Stderr = pwStderr
	cmd.Dir = workingDirectory

	log.Printf("[DEBUG] shell script command old state: \"%v\"", state)

	// Output what we're about to run
	log.Printf("[DEBUG] shell script going to execute: %s %s", shell, flag)
	commandLines := strings.Split(command, "\n")
	for _, line := range commandLines {
		log.Printf("   %s", line)
	}

	//send sdout and stderr to reader
	logCh := make(chan string)
	stdoutDoneCh := make(chan string)
	stderrDoneCh := make(chan string)
	go readOutput(prStderr, logCh, stderrDoneCh)
	go readOutput(prStdout, logCh, stdoutDoneCh)
	go logOutput(logCh)
	// Start the command
	log.Printf("-------------------------")
	log.Printf("[DEBUG] Starting execution...")
	log.Printf("-------------------------")
	err = cmd.Start()
	if err == nil {
		err = cmd.Wait()
	}

	// Close the write-end of the pipe so that the goroutine mirroring output
	// ends properly.
	pwStdout.Close()
	pwStderr.Close()
	stdOutput := <-stdoutDoneCh
	<-stderrDoneCh
	close(logCh)

	log.Printf("-------------------------")
	log.Printf("[DEBUG] Command execution completed:")
	log.Printf("-------------------------")
	o := getOutputMap(stdOutput)
	//if output is nil then no state was returned from output
	if o == nil {
		return nil, nil
	}
	newState := NewState(environment, o)
	log.Printf("[DEBUG] shell script command new state: \"%v\"", newState)
	return newState, nil
}

func logOutput(logCh chan string) {
	for line := range logCh {
		log.Printf("  %s", line)
	}
}

func readOutput(r io.Reader, logCh chan<- string, doneCh chan<- string) {
	defer close(doneCh)
	lr := linereader.New(r)
	output := ""
	for line := range lr.Ch {
		logCh <- line
		output += line
	}
	doneCh <- output
}

func parseJSON(s string) (map[string]string, error) {
	tb := bytes.Trim([]byte(s), "\x00")
	ts := string(tb)
	var f map[string]interface{}
	err := json.Unmarshal([]byte(ts), &f)
	if err != nil {
		log.Printf("[DEBUG] Unable to unmarshall data to map[string]string: '%v'", err)
		return nil, err
	}
	output := make(map[string]string)
	for k, v := range f {
		outputString, ok := v.(string)
		if !ok {
			outputString = ""
		}
		output[k] = outputString
	}
	return output, err
}

func getOutputMap(s string) map[string]string {
	//expect that the end of the output will have a JSON string that can be converted to a map[string]string
	re := regexp.MustCompile(`(?mU){.*}`)
	j := re.FindAllString(s, -1)
	log.Printf("[DEBUG] JSON string found: \n%s", j)
	m, err := parseJSON(j[len(j)-1])
	if err != nil {
		return nil
	} else {
		log.Printf("[DEBUG] Valid map[string]string:\n %v", m)
	}
	return m
}

func readFile(r io.Reader) string {
	buffer := new(bytes.Buffer)
	for {
		tmpdata := make([]byte, maxBufSize)
		bytecount, _ := r.Read(tmpdata)
		if bytecount == 0 {
			break
		}
		buffer.Write(tmpdata)
	}
	return buffer.String()
}
