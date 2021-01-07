resource "aws_key_pair" "kp" {
  key_name_prefix   = "${var.name}-kp"
  public_key = file(var.rsa_pub_path)
  tags = {
    resource_group = var.name
  }
}

module "ec2" {
  source            = "./modules/ec2"
  name              = var.name
  instance_count    = var.instance_count
  root_volume_size  = var.root_volume_size
  use_public_ip     = var.use_public_ip
  nat_gateway_count = var.nat_gateway_count
  subnets           = var.subnets
  region            = var.region
  key_name          = aws_key_pair.kp.key_name
  os                = var.os
  
  windows_instance_ami   = var.windows_instance_ami
  windows_instance_count = var.windows_instance_count

  providers = {
    aws = aws
  }
}
