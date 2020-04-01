# Also requires the following environment variables to be set:
#   DEPLOY_ENDPOINT = the domain name of the api
#   ENCRYPTION_KEY = the encryption key to use for creating the jwt

resource "shell_script" "pipeline" {
  lifecycle_commands {
    create = templatefile("${path.module}/scripts/create.sh", { STATE_FILE = "${var.name}.json" })
    read   = templatefile("${path.module}/scripts/read.sh", { STATE_FILE = "${var.name}.json" })
    update = templatefile("${path.module}/scripts/update.sh", { STATE_FILE = "${var.name}.json" })
    delete = templatefile("${path.module}/scripts/delete.sh", { STATE_FILE = "${var.name}.json" })
  }

  environment = {
    ORGANIZATION = var.organization
    NAME         = var.name
    DESCRIPTION  = var.description
  }
}
