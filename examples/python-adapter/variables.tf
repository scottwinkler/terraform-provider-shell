variable "name" {
  type = string
}

variable "organization" {
  type = string
}

variable "description" {
  default = ""
  type    = string
}

variable "stages" {
  default = {
    dev  = null
    qa   = null
    int  = null
    stg  = null
    prod = null
  }
  type = any
}