output "private_ip_lin" {
  value = module.ec2.private_ip_lin
}

output "public_ip_lin" {
  value = module.ec2.public_ip_lin
}

output "private_ip_win" {
  value = module.ec2.private_ip_win
}

output "public_ip_win" {
  value = module.ec2.public_ip_win
}

output "vpc_id" {
  value = module.ec2.vpc_id
}

output "public_subnet_ids" {
  value = module.ec2.public_subnet_ids
}

output "private_subnet_ids" {
  value = module.ec2.private_subnet_ids
}

output "private_route_table_id" {
  value = module.ec2.private_route_table
}
