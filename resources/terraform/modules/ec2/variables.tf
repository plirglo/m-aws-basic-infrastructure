variable "name" {
  description = "Name to be used on all resources as prefix"
  type        = string
  default     = "awsbi"
}

variable "instance_type" {
  description = "The type of instance to start"
  type        = string
  default     = "t3.medium"
}

variable "instance_count" {
  description = "Number of instances to launch"
  type        = number
}

variable "root_volume_size" {
  description = "The size of the root volume in gibibytes (GiB)"
  type        = number
}

variable "use_public_ip" {
  description = "If true, the EC2 instance will have associated public IP address"
  type        = bool
}

variable "nat_gateway_count" {
  description = "The number of nat gateways to create"
  type        = number
}

variable "region" {
  description = "Region to launch in"
  type        = string
}

variable "key_name" {
  description = "Key pair name"
  type        = string
}

variable "vpc_cidr_block" {
  description = "The cidr block of the VPC"
  default     = "10.1.0.0/20"
}

variable "subnets" {
  description = "Subnets configuration"
  type = object({
    private = object({
      count = number
    })
    public = object({
      count = number
    })
  })
  validation {
    condition     = (var.subnets.private.count > 0 && var.subnets.public.count > 0) || var.subnets.public.count > 0
    error_message = "At least one subnet should be created."
  }
}

variable "os" {
  description = "Operating System to launch"
  type = string
}

variable "windows_instance_ami" {
  description = "Operating System to launch on Windows nodes"
  type = string
}

variable "windows_instance_count" {
  description = "Number of Windows instances to launch"
  type        = number
}