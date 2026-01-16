# data "terraform_remote_state" "network" {
#   backend = "s3"
#   config = {
#     bucket = var.iacconsole_backend_config.bucket
#     key    = "network/terraform.tfstate"
#     region = var.iacconsole_backend_config.region
#   }
# }