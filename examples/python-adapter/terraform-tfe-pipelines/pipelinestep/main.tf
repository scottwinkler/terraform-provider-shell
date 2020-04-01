# Also requires the following environment variables to be set:
#   DEPLOY_ENDPOINT = the domain name of the api
#   ENCRYPTION_KEY = the encryption key to use for creating the jwt

resource "shell_script" "pipeline_step" {
  count = var.enabled ? 1 : 0
  lifecycle_commands {
    create = templatefile("${path.module}/scripts/create.sh",{STATE_FILE="${var.workspace_id}.json"})
    read   = templatefile("${path.module}/scripts/read.sh",{STATE_FILE="${var.workspace_id}.json"})
    update = templatefile("${path.module}/scripts/update.sh",{STATE_FILE="${var.workspace_id}.json"})
    delete = templatefile("${path.module}/scripts/delete.sh",{STATE_FILE="${var.workspace_id}.json"})
  }

  environment = {
    NAME         = var.name
    PIPELINE_ID  = var.pipeline_id
    DESCRIPTION  = var.description
    WORKSPACE_ID = var.workspace_id
    NEXT         = join(",", var.next)
    APPROVERS    = join(",", var.approvers)
    ENABLED      = var.enabled
  }
}
