output "id" {
    value = var.enabled? shell_script.pipeline_step[0].output["id"] : ""
}
