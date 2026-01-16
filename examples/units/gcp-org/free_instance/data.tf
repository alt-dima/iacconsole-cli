# data "terraform_remote_state" "free_instance" {
#   backend = "gcs"
#   config = {
#     bucket  = var.iacconsole_backend_config.bucket
#     prefix  = "account_free-tier/free_instance.tfstate"
#   }
# }