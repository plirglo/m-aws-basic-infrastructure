ROOT_DIR := $(patsubst %/,%,$(dir $(abspath $(firstword $(MAKEFILE_LIST)))))

AWSBI_IMAGE_NAME := epiphanyplatform/awsbi:0.0.1

-include $(ROOT_DIR)/vars.mk

M_NAME ?= awsbi-module
M_VMS_RSA_FILENAME ?= vms_rsa

export

.PHONY: all metadata setup init plan apply output all-destroy destroy-plan destroy clean

warning:
	$(error Usage: make (all/metadata/setup/init/plan/apply/output/all-destroy/destroy/destroy-plan/clean) )

all: init plan apply
all-destroy: destroy-plan destroy

clean:
	@rm -rf $(ROOT_DIR)/shared

metadata: setup
	@cd $(M_WORKDIR) && $(MAKE) metadata

init: setup
	@cd $(M_WORKDIR) && $(MAKE) init \
		M_NAME=$(M_NAME)

plan: setup
	@cd $(M_WORKDIR) && $(MAKE) plan \
		M_AWS_ACCESS_KEY=$(AWS_ACCESS_KEY) \
		M_AWS_SECRET_KEY=$(AWS_SECRET_KEY)

apply: setup
	@cd $(M_WORKDIR) && $(MAKE) apply \
		M_AWS_ACCESS_KEY=$(AWS_ACCESS_KEY) \
		M_AWS_SECRET_KEY=$(AWS_SECRET_KEY)

output: setup
	@cd $(M_WORKDIR) && $(MAKE) output \
		M_AWS_ACCESS_KEY=$(AWS_ACCESS_KEY) \
		M_AWS_SECRET_KEY=$(AWS_SECRET_KEY)

destroy-plan: setup
	@cd $(M_WORKDIR) && $(MAKE) plan-destroy \
		M_AWS_ACCESS_KEY=$(AWS_ACCESS_KEY) \
		M_AWS_SECRET_KEY=$(AWS_SECRET_KEY)

destroy: setup
	@cd $(M_WORKDIR) && $(MAKE) destroy \
		M_AWS_ACCESS_KEY=$(AWS_ACCESS_KEY) \
		M_AWS_SECRET_KEY=$(AWS_SECRET_KEY)

setup: $(M_SHARED)/$(M_VMS_RSA_FILENAME)

$(M_SHARED)/$(M_VMS_RSA_FILENAME):
	@mkdir -p $(dir $@)
	@ssh-keygen -t rsa -b 4096 -f $@ -N '' 2>&1 >/dev/null
