resource "aws_vpc" "example" {
  cidr_block = var.cidr 
  enable_dns_support = var.enable_dns_support
}