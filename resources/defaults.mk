define _M_SUBNETS
{
  private: {
    count: 1
  },
  public: {
    count: 1
  }
}
endef

M_VMS_COUNT ?= 1
M_PUBLIC_IPS ?= false
M_NAT_GATEWAY_COUNT ?= 1
M_SUBNETS ?= $(_M_SUBNETS)
M_REGION ?= eu-central-1
M_NAME ?= epiphany
M_VMS_RSA ?= vms_rsa
M_OS ?= redhat
M_WIN_AMI ?= ami-0c51aabac3acf27fe
M_WIN_COUNT ?= 0

AWS_ACCESS_KEY_ID ?= unset
AWS_SECRET_ACCESS_KEY ?= unset
