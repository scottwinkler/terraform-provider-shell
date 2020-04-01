# Also requires the following environment variables to be set:
#   ADDRESS = the FQDN of TFE
#   TOKEN = the TFE admin token
data "shell_script" "organizations" {
  lifecycle_commands {
    read = file("${path.module}/scripts/read.sh")
  }
  working_directory = path.module
}