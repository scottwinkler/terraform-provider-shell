package shell

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
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

func testAccDataShellScriptConfig(outValue string) string {
	return fmt.Sprintf(`
	data "shell_script" "test" {
	  lifecycle_commands {
		read = <<EOF
		  echo '{"out1": "%s"}' >&3
EOF
	  }
	}
`, outValue)
}
