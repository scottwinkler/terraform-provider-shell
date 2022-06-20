package shell

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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

func TestAccShellShellScript_failedUpdate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccShellScriptConfig_failedUpdate("value1"),
				Check:  resource.TestCheckResourceAttr("shell_script.shell_script", "environment.VALUE", "value1"),
			},
			{
				Config:             testAccShellScriptConfig_failedUpdate("value2"),
				ExpectNonEmptyPlan: true,
				ExpectError:        regexp.MustCompile("Error occured during shell execution"),
				Check:              resource.TestCheckResourceAttr("shell_script.shell_script", "environment.VALUE", "value1"),
			},
		},
	})
}

func testAccShellScriptConfig_failedUpdate(value string) string {
	return fmt.Sprintf(`
		resource "shell_script" "shell_script" {
			lifecycle_commands {
				create = "echo"
				read = <<-EOF
					echo -n '{"test": true}'
				EOF
				update = "exit 1"
				delete = "echo"
			}
			environment = {
				VALUE = "%s"
			}
		}
	`, value)
}

func testAccCheckNoFiles(files ...string) func(t *terraform.State) error {
	return func(t *terraform.State) error {
		for _, f := range files {
			if _, err := os.Stat(f); err == nil {
				return fmt.Errorf("'%s' should no longer exist", f)
			}
		}
		return nil
	}
}

func TestAccShellShellScript_recreate(t *testing.T) {
	file1, file2 := "/tmp/some-file-"+acctest.RandString(16), "/tmp/some-file-"+acctest.RandString(16)
	resource.Test(t, resource.TestCase{
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckNoFiles(file1, file2),
		Steps: []resource.TestStep{
			{
				Config: testAccShellShellScriptConfig_recreate(file1),
			},
			{
				Config: testAccShellShellScriptConfig_recreate(file2),
			},
		},
	})
}
func testAccShellShellScriptConfig_recreate(filename string) string {
	return fmt.Sprintf(`
		resource "shell_script" "shell_script" {
			lifecycle_commands {
				create = <<-EOF
					echo -n '{"test": true}' > "$FILE"
				EOF
				read = <<-EOF
					cat "$FILE"
				EOF
				delete = <<-EOF
					rm "$FILE"
				EOF
			}
			environment = {
				FILE = "%s"
			}
		}
	`, filename)
}

func TestAccShellShellScript_readFailed(t *testing.T) {
	file := "/tmp/test-file-" + acctest.RandString(16)
	resource.Test(t, resource.TestCase{
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckNoFiles(file),
		Steps: []resource.TestStep{
			{
				Config:             testAccShellShellScriptConfig_readFailed(file, true),
				ExpectNonEmptyPlan: true,
				Check:              resource.TestCheckResourceAttr("shell_script.shell_script", "output.test", "true"),
			},
			{
				Config: testAccShellShellScriptConfig_readFailed(file, false),
				Check:  resource.TestCheckResourceAttr("shell_script.shell_script", "output.test", "true"),
			},
		},
	})
}
func testAccShellShellScriptConfig_readFailed(filename string, bug bool) string {
	return fmt.Sprintf(`
		resource "shell_script" "shell_script" {
			lifecycle_commands {
				create = <<-EOF
					echo -n '{"test": true}' > "$FILE"
				EOF
				read = <<-EOF
					{ cat "$FILE"; [ "$BUG" == "true" ] && rm "$FILE" || true ;}
				EOF
				delete = <<-EOF
					rm "$FILE"
				EOF
			}
			environment = {
				FILE = "%s"
				BUG = "%t"
			}
		}
	`, filename, bug)
}

func TestAccShellShellScript_updateCommands(t *testing.T) {
	file := "/tmp/test-file-" + acctest.RandString(16)
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:             testAccShellShellScriptConfig_updateCommands(file, true),
				Check:              resource.TestCheckResourceAttr("shell_script.shell_script", "output.bug", "false"),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccShellShellScriptConfig_updateCommands(file, false),
				Check:  resource.TestCheckResourceAttr("shell_script.shell_script", "output.bug", "false"),
			},
			{
				Config: testAccShellShellScriptConfig_updateCommands(file, false),
				Check:  resource.TestCheckResourceAttr("shell_script.shell_script", "output.bug", "false"),
			},
		},
	})
}
func testAccShellShellScriptConfig_updateCommands(filename string, bug bool) string {
	var read = `cat "$FILE"`
	if bug {
		read = `[ -f "$FILE.bug" ] && cat "$FILE.bug" || { cat "$FILE" ; echo -n '{}' > "$FILE.bug" ;}`
	}

	return fmt.Sprintf(`
		resource "shell_script" "shell_script" {
			lifecycle_commands {
				create = <<-EOF
					echo -n '{"bug": false}' > "$FILE"
				EOF
				read = <<-EOF
					%s
				EOF
				update = <<-EOF
					echo -n '{"bug": true}' > "$FILE"
				EOF
				delete = <<-EOF
					rm "$FILE"
				EOF
			}
			environment = {
				FILE = "%s"
			}
		}
	`, read, filename)
}

func TestAccShellShellScript_outputDependency(t *testing.T) {
	file := "/tmp/test-file-" + acctest.RandString(16)
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccShellShellScriptConfig_outputDependency(file, true, false),
			},
			{
				Config: testAccShellShellScriptConfig_outputDependency(file, false, true),
				Check:  resource.TestCheckResourceAttr("shell_script.dependent", "triggers.output", "false"),
			},
		},
	})

}

func testAccShellShellScriptConfig_outputDependency(filename string, value bool, wdependent bool) (conf string) {
	conf = fmt.Sprintf(`
		resource "shell_script" "shell_script" {
			lifecycle_commands {
				create = <<-EOF
					echo -n '{"value": '"$VALUE"'}' > "$FILE"
				EOF
				read = <<-EOF
					cat "$FILE"
				EOF
				update = <<-EOF
					echo -n '{"value": '"$VALUE"'}' > "$FILE"
				EOF
				delete = <<-EOF
					rm "$FILE"
				EOF
			}
			environment = {
				FILE = "%s"
				VALUE = "%t"
			}
		}
	`, filename, value)

	if wdependent {
		conf = conf + `
			resource "shell_script" "dependent" {
				lifecycle_commands {
					create = "echo -n"
					read = "echo -n '{}'"
					delete = "echo -n"
				}

				triggers = {
					output = shell_script.shell_script.output["value"]
				}
			}
		`
	}
	return
}

func TestAccShellShellScript_previousOutput(t *testing.T) {
	file := "/tmp/test-file-" + acctest.RandString(16)
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccShellShellScriptConfig_previousOutput(file, true),
				Check:  resource.TestCheckResourceAttr("shell_script.previous", "output.value", ""),
			},
			{
				Config: testAccShellShellScriptConfig_previousOutput(file, false),
				Check:  resource.TestCheckResourceAttr("shell_script.previous", "output.value", "true"),
			},
		},
	})
}

func testAccShellShellScriptConfig_previousOutput(filename string, value bool) (conf string) {
	return fmt.Sprintf(`
		resource "shell_script" "shell_script" {
			lifecycle_commands {
				create = <<-EOF
					echo -n '{"value": '"$VALUE"'}' > "$FILE"
				EOF
				read = <<-EOF
					cat "$FILE"
				EOF
				update = <<-EOF
					cat > "$FILE.previous" && echo -n '{"value": '"$VALUE"'}' > "$FILE"
				EOF
				delete = <<-EOF
					rm "$FILE"
				EOF
			}
			environment = {
				FILE = "%s"
				VALUE = "%t"
			}
		}

		resource "shell_script" "previous" {
			lifecycle_commands {
				create = "true"
				read = <<-EOF
					cat $FILE || echo -n '{"value": null}'
				EOF
				update = "true"
				delete = "true"
			}
			environment = {
				FILE = "${shell_script.shell_script.environment.FILE}.previous"
			}
			triggers = {
				VALUE = shell_script.shell_script.environment.VALUE
			}
		}
	`, filename, value)
}
