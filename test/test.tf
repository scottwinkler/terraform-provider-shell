provider "shell" {}

data "shell_script" "test" {
  lifecycle_commands {
    read = <<EOF
      echo '{"commit_id": "b8f2b8b"}' >&3
    EOF
  }
}

output "commit_id" {
  value = "${data.shell_script.test.output["commit_id"]}"
}
/*
resource "shell_script" "test" {
  lifecycle_commands {
    create = "bash create.sh"
    read   = "bash read.sh"
    delete = "bash delete.sh"
  }

  working_directory = "./scripts"

  environment = {
    yolo = "yolo"
  }
}

output "environment" {
  value = "${shell_script.test.output["environment"]}"
}
*/