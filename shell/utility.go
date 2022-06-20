package shell

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/tidwall/gjson"
)

// CommandConfig is a helper struct
type CommandConfig struct {
	Command              string
	Environment          []string
	SensitiveEnvironment []string
	Interpreter          []string
	WorkingDirectory     string
	Action               Action
	PreviousOutput       map[string]string
	EnableParallelism    bool
}

func runCommand(c *CommandConfig) (map[string]string, error) {
	if !c.EnableParallelism {
		log.Printf("[DEBUG] parallelism disabled, locking")
		shellMutex.Lock()
		defer shellMutex.Unlock()
		log.Printf("[DEBUG] lock acquired")
	}

	// Setup the command
	shell := c.Interpreter[0]
	flags := append(c.Interpreter[1:], c.Command)
	cmd := exec.Command(shell, flags...)
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
	log.Printf("[DEBUG] shell script going to execute: %s %s", shell, flags)
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
	secrets := expandEnvironmentVariables(c.SensitiveEnvironment)
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
	stdError := <-stderrDoneCh
	close(logCh)

	// If the script exited with a non-zero code then send the error up to Terraform
	if err != nil {
		errorS := "Error occured during shell execution.\n"
		errorS += "Error: \n" + err.Error() + "\n\n"
		errorS += "Command: \n" + sanitizeString(c.Command, secretValues) + "\n\n"
		errorS += "StdOut: \n" + sanitizeString(stdOutput, secretValues) + "\n\n"
		errorS += "StdErr: \n" + sanitizeString(stdError, secretValues) + "\n\n"
		errorS += fmt.Sprintf("Env: \n%s\n\n", c.Environment)
		if c.Action != ActionCreate {
			stdin, _ := json.Marshal(c.PreviousOutput)
			errorS += fmt.Sprintf("StdIn: \n'%s'\n", stdin)
		}
		return nil, errors.New(errorS)
	}

	log.Printf("-------------------------")
	log.Printf("[DEBUG] Command execution completed:")
	log.Printf("-------------------------")
	o := getOutputMap(stdOutput, secretValues)
	return o, nil
}

func parseJSON(s string, secretValues []string) (map[string]string, error) {
	if !gjson.Valid(s) {
		return nil, fmt.Errorf("Invalid JSON: %s", sanitizeString(s, secretValues))
	}
	output := make(map[string]string)
	result := gjson.Parse(s)
	result.ForEach(func(key, value gjson.Result) bool {
		output[key.String()] = value.String()
		return true
	})
	return output, nil
}

func getOutputMap(s string, secretValues []string) map[string]string {
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
		m, err = parseJSON(match, secretValues)
		if err == nil {
			//match found
			break
		}
	}

	if m == nil {
		log.Printf("[DEBUG] no valid JSON strings found at end of output: \n%s", sanitizeString(s, secretValues))
		return nil
	}

	// be extra careful to not print anything we shouldn't (even if it shows up in state file anyways)
	sanitizedM := make(map[string]string)
	for k, v := range m {
		sanitizedM[k] = sanitizeString(v, secretValues)
	}
	log.Printf("[DEBUG] Valid map[string]string:\n %v", sanitizedM)
	return m
}

func expandEnvironmentVariables(envList []string) map[string]string {
	envMap := make(map[string]string)
	for _, v := range envList {
		parts := strings.Split(v, "=")
		key := parts[0]
		value := parts[1]
		envMap[key] = value
	}
	return envMap
}
