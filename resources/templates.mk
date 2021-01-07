define M_METADATA_CONTENT
labels:
  version: $(M_VERSION)
  name: AWS Basic Infrastructure
  short: $(M_MODULE_SHORT)
  kind: infrastructure
  provider: aws
  provides-vms: true
  provides-pubips: $(M_PUBLIC_IPS)
endef

define M_CONFIG_CONTENT
kind: $(M_MODULE_SHORT)-config
$(M_MODULE_SHORT):
  name: $(M_NAME)
  instance_count: $(M_VMS_COUNT)
  region: $(M_REGION)
  use_public_ip: $(M_PUBLIC_IPS)
  nat_gateway_count: $(M_NAT_GATEWAY_COUNT)
  subnets: $(M_SUBNETS)
  rsa_pub_path: "$(M_SHARED)/$(M_VMS_RSA).pub"
  os: $(M_OS)
  windows_instance_ami: $(M_WIN_AMI)
  windows_instance_count: $(M_WIN_COUNT)
endef

define M_STATE_INITIAL
kind: state
$(M_MODULE_SHORT):
  status: initialized
endef
