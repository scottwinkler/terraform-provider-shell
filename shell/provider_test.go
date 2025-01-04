package shell

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/stretchr/testify/require"
)

var testAccProviders map[string]terraform.ResourceProvider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider().(*schema.Provider)
	testAccProvider.ConfigureFunc = testProviderConfigure

	testAccProviders = map[string]terraform.ResourceProvider{
		"shell": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().(*schema.Provider).InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ terraform.ResourceProvider = Provider()
}

func TestProvider_HasChildResources(t *testing.T) {
	expectedResources := []string{
		"shell_script",
		"shell_sensitive_script",
	}

	resources := testAccProvider.ResourcesMap
	require.Equal(t, len(expectedResources), len(resources), "There are an unexpected number of registered resources")

	for _, resource := range expectedResources {
		require.Contains(t, resources, resource, "An expected resource was not registered")
		require.NotNil(t, resources[resource], "A resource cannot have a nil schema")
	}
}

func TestProvider_HasChildDataSources(t *testing.T) {
	expectedDataSources := []string{
		"shell_script",
		"shell_sensitive_script",
	}

	dataSources := testAccProvider.DataSourcesMap
	require.Equal(t, len(expectedDataSources), len(dataSources), "There are an unexpected number of registered data sources")

	for _, resource := range expectedDataSources {
		require.Contains(t, dataSources, resource, "An expected data source was not registered")
		require.NotNil(t, dataSources[resource], "A data source cannot have a nil schema")
	}
}

func testProviderConfigure(d *schema.ResourceData) (interface{}, error) {
	config := Config{}
	config.Environment = map[string]interface{}{
		"TEST_ENV1": "Env1_Val01",
		"TEST_ENV2": "Env2_Val02",
	}

	return config.Client()
}
