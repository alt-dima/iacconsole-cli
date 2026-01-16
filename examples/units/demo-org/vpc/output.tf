output "region_from_env" {
    value = var.iacconsole_envvar_awsregion
}

output "region_from_inv" {
    value = var.iacconsole_account_data.region
}