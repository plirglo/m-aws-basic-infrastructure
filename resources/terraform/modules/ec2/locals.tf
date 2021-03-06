locals {
  use_nat_gateway           = var.nat_gateway_count > 0
  select_ami                = var.os == "redhat" ? "RHEL-7.8_HVM_GA-20200225-x86_64-1-Hourly2-GP2" : "ubuntu/images/hvm-ssd/ubuntu-bionic-18.04-amd64-server-20200611"
  select_owner              = var.os == "redhat" ? "309956199498" : "099720109477"
  public_subnet_numbers     = range(var.subnets.public.count)
  private_subnet_numbers    = range(var.subnets.public.count, var.subnets.public.count + var.subnets.private.count)
  public_cidr_blocks        = [
    for num in local.public_subnet_numbers:
    cidrsubnet(var.vpc_cidr_block, 4, num)
  ]
  private_cidr_blocks       = [
    for num in local.private_subnet_numbers:
    cidrsubnet(var.vpc_cidr_block, 4, num)
  ]
}
