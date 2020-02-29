package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	r "githubdev.dco.elmae/CloudPlatform/tfe-module-registry/modules/golang/resources"
)

func getEnvironment() map[string]string {
	m := make(map[string]string)
	for _, e := range os.Environ() {
		if i := strings.Index(e, "="); i >= 0 {
			m[e[:i]] = e[i+1:]
		}
	}
	return m
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func initializeResource(name string, path string) (interface{}, error) {
	var resource interface{}
	if constructor, ok := r.DataResourceMap[name]; ok {
		fmt.Printf("Initializing new data resource: %s\n", name)
		environment := getEnvironment()
		resource = constructor(environment)
	} else if constructor, ok := r.ResourceMap[name]; ok {
		fmt.Printf("Initializing new resource: %s\n", name)
		environment := getEnvironment()
		state := readState(path)
		resource = constructor(environment, state)
	} else {
		err := fmt.Errorf("resource: %s does not exist in either the DataResourceMap or ResourceMap", name)
		return nil, err
	}
	return resource, nil
}

func executeCommand(command string, resource interface{}) map[string]string {
	var state map[string]string
	if t, ok := resource.(r.IResource); ok {
		switch command {
		case "create":
			t.Create()
		case "read":
			t.Read()
		case "update":
			t.Update()
		case "delete":
			t.Delete()
		default:
			fmt.Printf("Command not one of: ['create','read','update','delete']. Unexpected behavior may result\n")
		}
		state = t.State()
	} else if t, ok := resource.(r.IDataResource); ok {
		switch command {
		case "read":
			t.Read()
		default:
			fmt.Printf("Command not one of: ['read']. Unexpected behavior may result\n")
		}
		state = t.State()
	}
	return state
}

func readState(path string) map[string]string {
	dat, _ := ioutil.ReadFile(path)
	//check(err)
	m := make(map[string]string)
	json.Unmarshal(dat, &m)
	fmt.Printf("Old state: %v", m)
	return m
}

func saveState(state map[string]string, path string) {
	fmt.Printf("saving state: |%v| to path: %s\n", state, path)
	dat, _ := json.Marshal(state)
	ioutil.WriteFile(path, dat, 0755)
}

//example usage: ./golang -name=<name> -command=<command> -state=<path/to/state>
func main() {
	namePtr := flag.String("name", "", "name")
	commandPtr := flag.String("command", "", "command")
	pathPtr := flag.String("state", "state.json", "state")
	flag.Parse()
	name := *namePtr
	command := *commandPtr
	path := *pathPtr
	resource, err := initializeResource(name, path)
	check(err)
	state := executeCommand(command, resource)
	saveState(state, path)
}
