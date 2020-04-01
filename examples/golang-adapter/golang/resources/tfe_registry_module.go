package resources

import (
	"context"
	"fmt"
	"strings"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/tidwall/gjson"
)

const (
	POST = "POST"
	GET  = "GET"
)

type TFE_RegistryModule struct {
	environment map[string]string
	state       map[string]string
	client      *tfe.Client
}

func NewTFERegistryModule(environment map[string]string, state map[string]string) IResource {
	address := environment["ADDRESS"]
	token := environment["TOKEN"]
	client := GetTFEClient(address, token)
	return &TFE_RegistryModule{
		environment: environment,
		state:       state,
		client:      client,
	}
}

func (r *TFE_RegistryModule) getOAuthTokenId(organization string) string {
	tfeClient := r.client
	options := tfe.OAuthTokenListOptions{
		ListOptions: tfe.ListOptions{
			PageNumber: 999,
			PageSize:   200,
		},
	}
	ctx := context.Background()
	oAuthTokenList, err := tfeClient.OAuthTokens.List(ctx, organization, options)
	if err != nil {
		fmt.Printf("An error has occured while reading oAuth Token: %v\n", err)
		panic("error")
	}
	oAuthTokenID := oAuthTokenList.Items[0].ID
	return oAuthTokenID
}

func (r *TFE_RegistryModule) Read() {
	fmt.Println("Reading...")
	address := r.environment["ADDRESS"]
	token := r.environment["TOKEN"]
	organizationsStr := r.environment["ORGANIZATIONS"]
	organizations := strings.Split(organizationsStr, ",")
	module := r.state["module"]
	provider := r.state["provider"]
	for i := 0; i < len(organizations); i++ {
		organization := organizations[i]
		endpoint := fmt.Sprintf("%s/api/registry/v1/modules/search?q=%s&provider=%s&namespace=%s", address, module, provider, organization)
		resp, _ := TFERegistryModuleRequest(endpoint, token, GET, nil)
		isCreated := len(gjson.Get(resp, "modules").Array()) > 0
		if !isCreated {
			state := make(map[string]string)
			r.state = state
			return
		}
	}
}

func (r *TFE_RegistryModule) Create() {
	fmt.Println("Creating...")
	address := r.environment["ADDRESS"]
	token := r.environment["TOKEN"]
	organizationsStr := r.environment["ORGANIZATIONS"]
	vcsRepoIdentifier := r.environment["VCS_REPO_IDENTIFIER"]
	organizations := strings.Split(organizationsStr, ",")
	endpoint := fmt.Sprintf("%s/api/v2/registry-modules", address)

	state := make(map[string]string)
	for i := 0; i < len(organizations); i++ {
		organization := organizations[i]
		oAuthTokenID := r.getOAuthTokenId(organization)
		s := fmt.Sprintf(`{"data": {"attributes": {"vcs-repo": {"identifier": "%s","oauth-token-id": "%s","branch":"","display-identifier":"%s","github-app-installation-id":null,"ingress-submodules":true,"webhook-url":""}},"type": "registry-modules"}}`,
			vcsRepoIdentifier, oAuthTokenID, vcsRepoIdentifier)
		fmt.Printf("Payload: %v\n", s)
		var payload = []byte(s)
		resp, err := TFERegistryModuleRequest(endpoint, token, POST, payload)
		if err != nil {
			fmt.Printf("An error has occured while creating: %v\n attempting to Create()...\n", err)
			panic("error")
		}
		module := gjson.Get(resp, "data.attributes.name").String()
		provider := gjson.Get(resp, "data.attributes.provider").String()
		state["module"] = module
		state["provider"] = provider
	}
	r.state = state
	fmt.Printf("state: |%v|\nt", r.state)
}

func (r *TFE_RegistryModule) Update() {
	fmt.Println("Updating...")
	r.Delete()
	r.Create()
}

func (r *TFE_RegistryModule) Delete() {
	fmt.Println("Deleting...")
	address := r.environment["ADDRESS"]
	token := r.environment["TOKEN"]
	organizationsStr := r.environment["ORGANIZATIONS"]
	organizations := strings.Split(organizationsStr, ",")
	fmt.Printf("%v", r.state)
	module := r.state["module"]
	provider := r.state["provider"]
	fmt.Printf("module:%s,provider:%s", module, provider)
	for i := 0; i < len(organizations); i++ {
		organization := organizations[i]
		endpoint := fmt.Sprintf("%s/api/v2/registry-modules/actions/delete/%s/%s/%s", address, organization, module, provider)
		var payload []byte
		TFERegistryModuleRequest(endpoint, token, POST, payload)
	}
	state := make(map[string]string)
	r.state = state
}

func (r *TFE_RegistryModule) State() map[string]string {
	return r.state
}
