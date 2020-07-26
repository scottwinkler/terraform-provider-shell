package shell

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccShellShellScript_basic(t *testing.T) {
	resourceName := "shell_script.basic"
	rString := acctest.RandString(8)
	resource.Test(t, resource.TestCase{
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckShellScriptDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccShellScriptConfig_basic(rString),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "output.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "output.out1", rString),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
		},
	})
}

func TestAccShellShellScript_basic_error(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckShellScriptDestroy,
		Steps: []resource.TestStep{
			{
				Config:             testAccShellScriptConfig_basic_error(),
				ExpectNonEmptyPlan: true,
				ExpectError:        regexp.MustCompile("Something went wrong!"),
			},
		},
	})
}

func TestAccShellShellScript_basic_tagged_error(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckShellScriptDestroy,
		Steps: []resource.TestStep{
			{
				Config:             testAccShellScriptConfig_basic_tagged_error(),
				ExpectNonEmptyPlan: true,
				ExpectError:        regexp.MustCompile("Tag: testErrorTag"),
			},
		},
	})
}

func TestAccShellShellScript_create_read_delete(t *testing.T) {
	resourceName := "shell_script.crd"
	rString := acctest.RandString(8)
	resource.Test(t, resource.TestCase{
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckShellScriptDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccShellScriptConfig_create_read_delete(rString),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "output.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "output.out1", rString),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
		},
	})
}

func TestAccShellShellScript_create_update_delete(t *testing.T) {
	resourceName := "shell_script.cud"
	rString := acctest.RandString(8)
	resource.Test(t, resource.TestCase{
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckShellScriptDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccShellScriptConfig_create_update_delete(rString),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "output.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "output.out1", rString),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
		},
	})
}

func TestAccShellShellScript_complete(t *testing.T) {
	resourceName := "shell_script.complete"
	rString := acctest.RandString(8)
	resource.Test(t, resource.TestCase{
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckShellScriptDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccShellScriptConfig_complete(rString),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "output.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "output.out1", rString),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
		},
	})
}

func TestAccShellShellScript_providerEnvCud(t *testing.T) {
	resourceName := "shell_script.cud"
	resource.Test(t, resource.TestCase{
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckShellScriptDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccShellScriptConfigWithProviderEnv_create_update_delete(),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "output.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "output.value", "Env2_Val02"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
		},
	})
}

func testAccCheckShellScriptDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "shell_script" {
			continue
		}
		fileName := rs.Primary.Attributes["shell_script.test.environment.filename"]
		if _, err := os.Stat(fileName); os.IsExist(err) {
			return fmt.Errorf("Shell Script file failed to cleanup")
		}
	}
	return nil
}

func testAccShellScriptConfig_basic(outValue string) string {
	return fmt.Sprintf(`
	resource "shell_script" "basic" {
		lifecycle_commands {
		  create = <<EOF
			out='{"out1": "%s"}'
			touch create_delete.json
			echo $out >> create_delete.json
			cat create_delete.json
EOF
		  delete = "rm -rf create_delete.json"
		}
	  
		environment = {
		  filename= "create_delete.json"
		}
	  }
`, outValue)
}

func testAccShellScriptConfig_basic_error() string {
	return fmt.Sprintf(`
	resource "shell_script" "basic" {
		lifecycle_commands {
		  create = <<EOF
		    echo "Something went wrong!"
			exit 1
EOF
		  delete = "exit 1"
		}
	  
		environment = {
		  filename= "create_delete.json"
		}
	  }
`)
}

func testAccShellScriptConfig_basic_tagged_error() string {
	return fmt.Sprintf(`
	resource "shell_script" "basic" {
		lifecycle_commands {
		  create = <<EOF
		    echo "Something went wrong!"
			exit 1
EOF
		  delete = "exit 1"
		}

		environment = {
		  filename= "create_delete.json"
		}

		error_tag = "testErrorTag"
	  }
`)
}

func testAccShellScriptConfig_create_read_delete(outValue string) string {
	return fmt.Sprintf(`
	resource "shell_script" "crd" {
		lifecycle_commands {
		  create = <<EOF
			out='{"out1": "%s"}'
			touch create_read_delete.json
			echo $out >> create_read_delete.json
			cat create_read_delete.json
EOF
		  read   = "cat create_read_delete.json"
		  delete = "rm -rf create_read_delete.json"
		}

		environment = {
		  filename= "create_read_delete.json"
		}
	  }
`, outValue)
}

func testAccShellScriptConfig_create_update_delete(outValue string) string {
	return fmt.Sprintf(`
	resource "shell_script" "cud" {
		lifecycle_commands {
		  create = <<EOF
			out='{"out1": "%s"}'
			touch create_update_delete.json
			echo $out >> create_update_delete.json
			cat create_update_delete.json
EOF
		  update = <<EOF
			rm -rf create_update_delete.json
			out='{"out1": "%s"}'
			touch "create_update_delete.json"
			echo $out >> create_update_delete.json
			cat create_update_delete.json
EOF
		  delete = "rm -rf create_update_delete.json"
		}

		environment = {
			filename= "create_update_delete.json"
		}
	  }
`, outValue, outValue)
}

func testAccShellScriptConfig_complete(outValue string) string {
	return fmt.Sprintf(`
	resource "shell_script" "complete" {
		lifecycle_commands {
			create = file("test-fixtures/scripts/create.sh")
			read   = file("test-fixtures/scripts/read.sh")
			update = file("test-fixtures//scripts/update.sh")
			delete = file("test-fixtures/scripts/delete.sh")
		}

		environment = {
			filename= "create_complete.json"
			testdatasize = "100240"						
			out1 = "%s"
		}

		triggers = {
			key = "value"
		}
	  }
`, outValue)
}

func testAccShellScriptConfigWithProviderEnv_create_update_delete() string {
	return `
	resource "shell_script" "cud" {
		lifecycle_commands {
		  create = <<EOF
		    out="{\"value\": \"$TEST_ENV2\"}"
	        touch create_update_delete.json
			echo $out >> create_update_delete.json
			cat create_update_delete.json
EOF
		  update = <<EOF
			rm -rf create_update_delete.json
			out="{\"value\": \"$TEST_ENV2\"}"
			touch "create_update_delete.json"
			echo $out >> create_update_delete.json
			cat create_update_delete.json
EOF
		  delete = "rm -rf create_update_delete.json"
		}

		environment = {
			filename= "create_update_delete.json"
		}
	  }
`
}
