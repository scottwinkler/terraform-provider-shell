# Also requires the following environment variables to be set:
#   ADDRESS = the FQDN of TFE
#   TOKEN = the TFE admin token
resource "shell_script" "registry-module" {
  lifecycle_commands {
    create = file("${path.module}/scripts/create.sh")
    read   = file("${path.module}/scripts/read.sh")
    update = file("${path.module}/scripts/update.sh")
    delete = file("${path.module}/scripts/delete.sh")
  }
  working_directory = path.module
  environment = {
    VCS_REPO_IDENTIFIER = var.vcs_repo_identifier
    ORGANIZATIONS       = local.organizations
  }
}
