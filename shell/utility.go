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

	"github.com/armon/circbuf"
	"github.com/mitchellh/go-linereader"
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
		outputString, ok := v.(string)
		if !ok {
			outputString = ""
		}
		output[k] = outputString
	}
	return output, err
}

func runCommand(command string, state *State, environment []string, workingDirectory string) (*State, error) {
	shellMutexKV.Lock(shellScriptMutexKey)
	defer shellMutexKV.Unlock(shellScriptMutexKey)
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
	prDev3, pwDev3, err := os.Pipe()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize pipe for >&3 output: %s", err)
	}
	cmd.ExtraFiles = []*os.File{pwDev3}

	log.Printf("[DEBUG] shell script command old state: \"%v\"", state)

	// Output what we're about to run
	log.Printf("[DEBUG] shell script going to execute: %s %s", shell, flag)
	commandLines := strings.Split(command, "\n")
	for _, line := range commandLines {
		log.Printf("   %s", line)
	}
	// Write everything we read from the pipe to the output buffer too
	output, _ := circbuf.NewBuffer(maxBufSize)
	teeStdout := io.TeeReader(prStdout, output)
	teeStderr := io.TeeReader(prStderr, output)
	// copy the teed output to the UI output
	copyDoneCh := make(chan struct{})
	go copyOutput(io.MultiReader(teeStdout, teeStderr), copyDoneStdoutCh)

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
	pwDev3.Close()

	// Cancelling the command may block the pipe reader if the file descriptor
	// was passed to a child process which hasn't closed it. In this case the
	// copyOutput goroutine will just hang out until exit.
	select {
	case <-copyDoneCh:
	}
	log.Printf("-------------------------")
	log.Printf("[DEBUG] Command execution completed. Reading from output pipe: >&3")
	log.Printf("-------------------------")

	//read back diff output from pipe
	buffer := new(bytes.Buffer)
	for {
		tmpdata := make([]byte, maxBufSize)
		bytecount, _ := prDev3.Read(tmpdata)
		if bytecount == 0 {
			break
		}
		buffer.Write(tmpdata)
	}

	log.Printf("[DEBUG] shell script command output:\n%s", buffer.String())

	if err != nil {
		return nil, fmt.Errorf("Error running command: '%v'", err)
	}

	o, err := parseJSON(buffer.Bytes())
	if err != nil {
		log.Printf("[DEBUG] Unable to unmarshall data to map[string]string: '%v'", err)
		return nil, nil
	}
	newState := NewState(environment, o)
	log.Printf("[DEBUG] shell script command new state: \"%v\"", newState)
	return newState, nil
}

func copyOutput(r io.Reader, doneCh chan<- struct{}) {
	defer close(doneCh)
	lr := linereader.New(r)
	for line := range lr.Ch {
		log.Printf("  %s\n", line)
	}
}
