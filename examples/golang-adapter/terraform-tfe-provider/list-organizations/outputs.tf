output "organizations" {
   value = keys("${data.shell_script.organizations.output}")
}

