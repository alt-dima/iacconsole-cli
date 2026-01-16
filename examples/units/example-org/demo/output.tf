output "account_region_from_inv" {
    value = var.iacconsole_account_data.region
}

output "account_id_from_inv" {
    value = var.iacconsole_account_data.account_id
}

output "datacenter_name_from_inv" {
    value = var.iacconsole_datacenter_name
}
	
output "datacenter_vpc_cidr_from_inv" {
    value = var.iacconsole_datacenter_data.vpc_cidr
}
	
output "datacenter_az_count_from_inv" {
    value = var.iacconsole_datacenter_data.az_count
}