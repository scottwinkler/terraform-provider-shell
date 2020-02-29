package resources

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"

	tfe "github.com/hashicorp/go-tfe"
)

var client *tfe.Client
var err error
var once sync.Once

func GetTFEClient(address string, token string) *tfe.Client {
	once.Do(func() {
		cfg := &tfe.Config{
			Address: address,
			Token:   token,
		}
		client, err = tfe.NewClient(cfg)
		if err != nil {
			fmt.Errorf("Error configuring tfe client: %v", err)
		}
		if client == nil {
			fmt.Errorf("Error configuring client. Is the environment variable 'ADDRESS' set to https://[domain_name]? Also is the environment variable 'TOKEN' set and valid?")
		}
	})
	return client
}

func TFERegistryModuleRequest(address string, token string, method string, payload []byte) (string, error) {
	fmt.Printf("Sending request to: %s\n", address)
	var body = bytes.NewBuffer(payload)
	req, err := http.NewRequest(method, address, body)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/vnd.api+json")
	if err != nil {
		log.Println(err)
		return "", err
	}
	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Println(err)
		return "", err
	}
	f, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return "", err
	}
	resp.Body.Close()
	if err != nil {
		log.Fatal(err)
		return "", err
	}
	j := string(f)
	fmt.Printf("response: %s\n", j)
	return j, err
}
