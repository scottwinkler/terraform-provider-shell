variable "enabled" {
  default     = true
  description = "Is this step enabled?"
}

variable "name" {
  type        = string
  description = "Name of pipelineStep"
}

variable "pipeline_id" {
  description = "The name of the organization which the pipelineStep is associated with"
  type        = string
}

variable "description" {
  description = "Additional details about pipeline"
  default     = ""
  type        = string
}

variable "workspace_id" {
  description = "The workspace id that the pipelineStep is associated with"
  type        = string
}

variable "next" {
  description = "A list of pipelineStepIds which are the nodes following this. Usually one, but could be null for end of linked list."
  default     = []
  type        = list(string)
}

variable "approvers" {
  description = "A list of approvers for this step, if any. Should not be set for first step"
  default     = []
  type        = list(string)
}
