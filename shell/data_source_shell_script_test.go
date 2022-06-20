package shell

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccShellDataShellScript_basic(t *testing.T) {
	resourceName := "data.shell_script.test"
	rString := acctest.RandString(8)
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataShellScriptConfig(rString),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "output.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "output.out1", rString),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
		},
	})
}

func TestAccShellDataShellScript_withEnv(t *testing.T) {
	resourceName := "data.shell_script.test"
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testDataSourceShellScriptWithEnv,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "environment.TOKEN", "valueX"),
					resource.TestCheckResourceAttr(resourceName, "output.value", "valueX"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
			{
				Config: testDataSourceShellScriptWithProviderEnv,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "output.value", "Env1_Val01"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
			{
				Config: testDataSourceShellScriptEnvOverridesProviderEnv,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "output.value", "override_val"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
		},
	})
}

func testAccDataShellScriptConfig(outValue string) string {
	return fmt.Sprintf(`
	data "shell_script" "test" {
	  lifecycle_commands {
		read = <<EOF
		  echo '{"out1": "%s"}'
EOF
	  }
	}
`, outValue)
}

var testDataSourceShellScriptWithEnv = `
data "shell_script" "test" {
	lifecycle_commands {
	read = <<EOF
		echo "{\"value\": \"$TOKEN\"}"
EOF
	}
	environment = {
		"TOKEN" = "valueX"
	}
}
`

var testDataSourceShellScriptWithProviderEnv = `
data "shell_script" "test" {
	lifecycle_commands {
	read = <<EOF
		echo "{\"value\": \"$TEST_ENV1\"}"
EOF
	}
}
`

var testDataSourceShellScriptEnvOverridesProviderEnv = `
data "shell_script" "test" {
	lifecycle_commands {
	read = <<EOF
		echo "{\"value\": \"$TEST_ENV1\"}"
EOF
	}
	environment = {
		"TEST_ENV1" = "override_val"
	}
}
`
