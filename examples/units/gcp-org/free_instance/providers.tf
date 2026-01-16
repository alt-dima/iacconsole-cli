# provider "aws" {
#     //region = var.iacconsole_account_manifest.region
#     region = var.iacconsole_envvar_awsregion
# }

provider "google" {
  project = var.iacconsole_account_data.project
  region  = var.iacconsole_account_data.region
  zone    = var.iacconsole_account_data.zone
}