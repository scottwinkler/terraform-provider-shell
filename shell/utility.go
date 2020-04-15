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

	"github.com/tidwall/gjson"
)

// CommandConfig is a helper struct
type CommandConfig struct {
	Command              string
	Environment          []string
	SensitiveEnvironment []string
	WorkingDirectory     string
	Action               Action
	PreviousOutput       map[string]string
}

func runCommand(c *CommandConfig) (map[string]string, error) {
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
	cmd := exec.Command(shell, flag, c.Command)
	if c.Action != ActionCreate {
		input, _ := json.Marshal(c.PreviousOutput)
		stdin := bytes.NewReader(input)
		cmd.Stdin = stdin
	}
	cmd.Env = append(append(os.Environ(), c.Environment...), c.SensitiveEnvironment...)
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
	cmd.Dir = c.WorkingDirectory

	// Output what we're about to run
	log.Printf("[DEBUG] shell script going to execute: %s %s", shell, flag)
	commandLines := strings.Split(c.Command, "\n")
	for _, line := range commandLines {
		log.Printf("   %s", line)
	}

	//send sdout and stderr to reader
	logCh := make(chan string)
	stdoutDoneCh := make(chan string)
	stderrDoneCh := make(chan string)
	go readOutput(prStderr, logCh, stderrDoneCh)
	go readOutput(prStdout, logCh, stdoutDoneCh)

	// get secret values (if any) to sanitize in logs
	secrets := unpackEnvironmentVariables(c.SensitiveEnvironment)
	secretValues := make([]string, 0, len(secrets))
	for _, v := range secrets {
		secretValues = append(secretValues, v)
	}
	go logOutput(logCh, secretValues)

	// Start the command
	log.Printf("-------------------------")
	log.Printf("[DEBUG] Starting execution...")
	log.Printf("-------------------------")
	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	err = cmd.Wait()

	// Close the write-end of the pipe so that the goroutine mirroring output
	// ends properly.
	pwStdout.Close()
	pwStderr.Close()

	stdOutput := <-stdoutDoneCh
	stdOutput = sanitizeString(stdOutput, secretValues)
	stdError := <-stderrDoneCh
	stdError = sanitizeString(stdError, secretValues)
	close(logCh)

	// If the script exited with a non-zero code then send the error up to Terraform
	if err != nil {
		return nil, fmt.Errorf("Error occurred during execution.\n Command: '%s' \n Error: '%s' \n StdOut: \n %s \n StdErr: \n %s", c.Command, err.Error(), stdOutput, stdError)
	}

	log.Printf("-------------------------")
	log.Printf("[DEBUG] Command execution completed:")
	log.Printf("-------------------------")
	o := getOutputMap(stdOutput)
	return o, nil
}

func parseJSON(s string) (map[string]string, error) {
	if !gjson.Valid(s) {
		return nil, fmt.Errorf("Invalid JSON: %s", s)
	}
	output := make(map[string]string)
	result := gjson.Parse(s)
	result.ForEach(func(key, value gjson.Result) bool {
		output[key.String()] = value.String()
		return true
	})
	return output, nil
}

func getOutputMap(s string) map[string]string {
	//Find all matches of "{(.*)/g" in output
	var matches []string
	substring := s
	idx := strings.Index(substring, "{")
	for idx != -1 {
		substring = substring[idx:]
		matches = append(matches, substring)
		if len(substring) > 0 {
			substring = substring[1:]
		}
		idx = strings.Index(substring, "{")
	}

	//Use last match that is a valid JSON
	var m map[string]string
	var err error
	for i := range matches {
		match := matches[len(matches)-1-i]
		m, err = parseJSON(match)
		if err == nil {
			//match found
			break
		}
	}

	if m == nil {
		log.Printf("[DEBUG] no valid JSON strings found at end of output: \n%s", s)
		return nil
	}

	log.Printf("[DEBUG] Valid map[string]string:\n %v", m)
	return m
}

func readFile(r io.Reader) string {
	const maxBufSize = 8 * 1024
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
