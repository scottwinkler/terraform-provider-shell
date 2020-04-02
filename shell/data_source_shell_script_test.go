package shell

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
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
					resource.TestCheckResourceAttr(resourceName, "output.%", "4"),
					resource.TestCheckResourceAttr(resourceName, "output.out1", rString),
					resource.TestCheckResourceAttr(resourceName, "output.out2", rString),
					resource.TestCheckResourceAttr(resourceName, "output.out3", rString),
					resource.TestCheckResourceAttr(resourceName, "output.out4", rString),
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
		  echo '{"out1": "'$OUTVALUE1'", "out2": "'$OUTVALUE2'", "out3": "'$OUTVALUE3'", "out4": "%s"}'
EOF
	  }

	  environment = {
		OUTVALUE1 = "%s"
		OUTVALUE3 = "will be replaced"
	  }

	  environment_sensitive = {
		OUTVALUE2 = "%s"
		OUTVALUE3 = "%s"
	  }
	}
`, outValue, outValue, outValue, outValue)
}

