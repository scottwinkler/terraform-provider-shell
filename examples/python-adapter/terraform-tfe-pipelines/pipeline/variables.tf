variable "name" {
  type        = string
  description = "Name of pipeline"
}

variable "organization" {
  type        = string
  description = "The name of the organization which the pipeline is associated with"
}

variable "description" {
  type        = string
  description = "Additional details about pipeline"
  default     = ""
}