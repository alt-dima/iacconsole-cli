//use shared-module
module "vpc" {
  source = "./shared-modules/create_vpc"
  cidr = var.iacconsole_datacenter_data[var.iacconsole_account_name].cidr
  enable_dns_support = try(var.iacconsole_datacenter_data[var.iacconsole_account_name].enable_dns_support, var.iacconsole_datacenter_defaults[var.iacconsole_account_name].enable_dns_support)
}

module "vpc_example_simple-vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "5.7.1"

  cidr = var.iacconsole_datacenter_data[var.iacconsole_account_name].cidr
}