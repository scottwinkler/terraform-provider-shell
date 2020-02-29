module "pipeline" {
  source       = "./terraform-tfe-pipelines/pipeline"
  name         = var.name
  organization = var.organization
  description  = var.description
}

# 1 = true
# 0 = false
locals {
  dev_paths = {
    "qa-true_int-true_stg-true"    = module.qa-step.id
    "qa-true_int-false_stg-true"   = module.qa-step.id
    "qa-true_int-false_stg-false"  = module.qa-step.id
    "qa-true_int-false_stg-true"   = module.qa-step.id
    "qa-true_int-true_stg-false"   = module.qa-step.id
    "qa-false_int-true_stg-true"   = module.int-step.id
    "qa-false_int-true_stg-false"  = module.int-step.id
    "qa-false_int-false_stg-true"  = module.stg-step.id
    "qa-false_int-false_stg-false" = module.prod-step.id
  }
  qa_paths = {
    "int-true_stg-true"   = module.int-step.id
    "int-true_stg-false"  = module.int-step.id
    "int-false_stg-true"  = module.stg-step.id
    "int-false_stg-false" = module.prod-step.id
  }
  int_paths = {
    "stg-true"  = module.stg-step.id
    "stg-false" = module.prod-step.id
  }
  enabled = {
    dev  = lookup(var.stages, "dev", null) != null ? true : false
    qa   = lookup(var.stages, "qa", null) != null ? true : false
    int  = lookup(var.stages, "int", null) != null ? true : false
    stg  = lookup(var.stages, "stg", null) != null ? true : false
    prod = lookup(var.stages, "prod", null) != null ? true : false
  }

  dev_next = local.enabled.dev ? lookup(local.dev_paths, "qa-${local.enabled.qa}_int-${local.enabled.int}_stg-${local.enabled.stg}") : null
  qa_next  = local.enabled.qa  ? lookup(local.qa_paths, "int-${local.enabled.int}_stg-${local.enabled.stg}") : null
  int_next = local.enabled.int ? lookup(local.int_paths, "stg-${local.enabled.stg}") : null
  stg_next = local.enabled.stg ? module.prod-step.id : null
}

data "tfe_workspace" "dev" {
  count        = local.enabled.dev ? 1 : 0
  name         = var.stages.dev.workspace_name
  organization = var.organization
}

module "dev-step" {
  source       = "./terraform-tfe-pipelines/pipelinestep"
  enabled      = local.enabled.dev
  name         = "${var.name}-dev"
  description  = lookup(lookup(var.stages, "dev", {}), "description", "${var.name} dev step")
  pipeline_id  = module.pipeline.id
  workspace_id = data.tfe_workspace.dev[0].external_id
  next         = [local.dev_next]
  approvers    = []
}

data "tfe_workspace" "qa" {
  count        = local.enabled.qa ? 1 : 0
  name         = var.stages.qa.workspace_name
  organization = var.organization
}

module "qa-step" {
  source       = "./terraform-tfe-pipelines/pipelinestep"
  enabled      = local.enabled.qa
  name         = "${var.name}-qa"
  description  = lookup(lookup(var.stages, "qa", {}), "description", "${var.name} qa step")
  pipeline_id  = module.pipeline.id
  workspace_id = data.tfe_workspace.qa[0].external_id
  next         = [local.qa_next]
  approvers    = lookup(lookup(var.stages, "qa", {}), "approvers", [])
}

data "tfe_workspace" "int" {
  count        = local.enabled.int ? 1 : 0
  name         = var.stages.int.workspace_name
  organization = var.organization
}

module "int-step" {
  source       = "./terraform-tfe-pipelines/pipelinestep"
  enabled      = local.enabled.int
  name         = "${var.name}-int"
  description  = lookup(lookup(var.stages, "int", {}), "description", "${var.name} int step")
  pipeline_id  = module.pipeline.id
  workspace_id = data.tfe_workspace.int[0].external_id
  next         = [local.int_next]
  approvers    = lookup(lookup(var.stages, "int", {}), "approvers", [])
}

data "tfe_workspace" "stg" {
  count        = local.enabled.stg ? 1 : 0
  name         = var.stages.stg.workspace_name
  organization = var.organization
}

module "stg-step" {
  source       = "./terraform-tfe-pipelines/pipelinestep"
  enabled      = local.enabled.stg
  name         = "${var.name}-stg"
  description  = lookup(lookup(var.stages, "stg", {}), "description", "${var.name} stg step")
  pipeline_id  = module.pipeline.id
  workspace_id = data.tfe_workspace.stg[0].external_id
  next         = [local.stg_next]
  approvers    = lookup(lookup(var.stages, "stg", {}), "approvers", [])
}

data "tfe_workspace" "prod" {
  count        = local.enabled.prod ? 1 : 0
  name         = var.stages.prod.workspace_name
  organization = var.organization
}

module "prod-step" {
  source       = "./terraform-tfe-pipelines/pipelinestep"
  enabled      = local.enabled.prod
  name         = "${var.name}-prod"
  description  = lookup(lookup(var.stages, "prod", {}), "description", "${var.name} prod step")
  pipeline_id  = module.pipeline.id
  workspace_id = data.tfe_workspace.prod[0].external_id
  approvers    = lookup(lookup(var.stages, "prod", {}), "approvers", [])
}