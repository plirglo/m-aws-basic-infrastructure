output "private_ip" {
  value = module.ec2.private_ip
}

output "public_ip" {
  value = module.ec2.public_ip
}
