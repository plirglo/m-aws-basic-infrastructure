variable "name" {
  description = "Name to be used on all resources as prefix"
  type        = string
}

variable "instance_count" {
  description = "Number of Linux instances to launch"
  type        = number
}

variable "root_volume_size" {
  description = "The size of the root volume in gibibytes (GiB)"
  type        = number
  default     = 64
}

variable "region" {
  description = "Region to launch in"
  type        = string
}

variable "use_public_ip" {
  description = "If true, the EC2 instance will have associated public IP address"
  type        = bool
}

variable "nat_gateway_count" {
  description = "The number of nat gateways to create"
  type        = number
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
}

variable "rsa_pub_path" {
  type = string
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
