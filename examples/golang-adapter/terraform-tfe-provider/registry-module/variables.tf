variable "vcs_repo_identifier" {
  type        = string
  description = "Specifies the repository from which to ingress the configuration. For most VCS providers, the format is <ORGANIZATION>/<REPO>"
}

variable "organizations" {
  type        = list(string)
  description = "A list of all organizations"
}

locals {
    organizations = join(",","${var.organizations}")
}