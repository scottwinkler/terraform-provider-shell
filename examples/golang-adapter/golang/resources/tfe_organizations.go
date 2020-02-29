package resources

import (
	"context"
	"fmt"

	tfe "github.com/hashicorp/go-tfe"
)

type TFE_Organizations struct {
	environment map[string]string
	state       map[string]string
	client      *tfe.Client
}

func NewTFEOrganizations(environment map[string]string) IDataResource {
	address := environment["ADDRESS"]
	token := environment["TOKEN"]
	client := GetTFEClient(address, token)
	return &TFE_Organizations{
		environment: environment,
		client:      client,
	}
}

func (d *TFE_Organizations) Read() {
	fmt.Println("Reading...")
	tfeClient := d.client
	options := tfe.OrganizationListOptions{
		ListOptions: tfe.ListOptions{
			PageNumber: 999,
			PageSize:   200,
		},
	}
	ctx := context.Background()
	organizationList, _ := tfeClient.Organizations.List(ctx, options)
	organizations := organizationList.Items
	state := make(map[string]string)
	for i := 0; i < len(organizations); i++ {
		organization := organizations[i].Name
		state[organization] = organization
		fmt.Printf("org: %s\n", organization)
	}
	d.state = state
	fmt.Printf("state: |%v|", d.state)
}

func (d *TFE_Organizations) State() map[string]string {
	return d.state
}
