# #comment two
# 

locals {
  hsh = {
    "first" = {}
    "second" = {}
  }
}
locals {
  ecs_cluster_settings = {
    "cluster_name" = data.terraform_remote_state.network.outputs.project.name
    xcluster_name = [{
      key_one = ["2", {
        key_two = "in_value"
      }, "3", { key_four = { key_five = "six" } }]
    }]
    another_attr = {
      another_key = dirname("something")
    }
  }
}
 
locals {
 ecs_cluster_settings = {
   "cluster_name" = data.terraform_remote_state.network.outputs.project.name
 }
 ec2_autoscaling_settings = {
   "enabled"         = true
   "min"             = 7
   "max"             = 10
   "ami"             = data.aws_ami.ecs.id
   "target_capacity" = 100
 }
}

locals {
  project_owner    = basename(dirname(dirname(path.cwd)))
  project_env_type = basename(dirname(path.cwd))
  project = {
    owner    = local.project_owner
    env_type = local.project_env_type
    name     = "${local.project_owner}-${local.project_env_type}"
    ansible  = "../../ansible/${local.project_env_type}"
  }
}

# comment one

variable "organization" {
  type        = string
# inside comment
  description = "TF Cloud Organization name"
  default     = "Spryker"
}

# comment three

data "template_file" "frontend_config" {
  template = file("${path.module}/frontend.json")
}

module "acm_primary" {
  source            = "git@github.com:spryker/tfcloud-modules.git//aws_acm?ref=v5.4.0"
  domains_and_zones = { for zone in distinct([for d in jsondecode(data.template_file.frontend_config.rendered) : d.zone]) : "*.${zone}" => zone }
}
