ROOT_DIR := $(patsubst %/,%,$(dir $(abspath $(firstword $(MAKEFILE_LIST)))))

AWSBI := epiphany/awsbi:0.0.1

M_NAME ?= awsbi-module-test
M_PUBLIC_IPS ?= false
M_VMS_COUNT ?= 1
VMS_RSA_FILENAME ?= vms_rsa
M_USERNAME ?= ec2-user
M_AMI ?= ami-02ab606eae7264892

-include $(ROOT_DIR)/vars.mk

export

.PHONY: all clean metadata init plan apply output destroy-plan destroy setup

warning:
	$(error Please specify make target: make (all/init/plan/apply/setup/destroy/destroy-plan/clean) )

all: init plan apply

clean:
	@rm -rf $(ROOT_DIR)/shared #TODO change root_dir to external dir

metadata: setup
	@docker run --rm \
		-v $(ROOT_DIR)/shared:/shared \
		-t $(AWSBI) \
		metadata \

init: setup
	@docker run --rm \
		-v $(ROOT_DIR)/shared:/shared \
		-t $(AWSBI) \
		init \
		M_VMS_COUNT=$(M_VMS_COUNT) \
		M_PUBLIC_IPS=$(M_PUBLIC_IPS) \
		M_NAME=$(M_NAME) \
		M_VMS_RSA=$(VMS_RSA_FILENAME) \
		M_USERNAME=$(M_USERNAME) \
		M_AMI=$(M_AMI)

plan: setup
	@docker run --rm \
		-v $(ROOT_DIR)/shared:/shared \
		-t $(AWSBI) \
		plan \
		M_AWS_ACCESS_KEY=$(AWS_ACCESS_KEY) \
		M_AWS_SECRET_KEY=$(AWS_SECRET_KEY)

apply: setup
	@docker run --rm \
		-v $(ROOT_DIR)/shared:/shared \
		-t $(AWSBI) \
		apply \
		M_AWS_ACCESS_KEY=$(AWS_ACCESS_KEY) \
		M_AWS_SECRET_KEY=$(AWS_SECRET_KEY)

output: setup
	@docker run --rm \
		-v $(ROOT_DIR)/shared:/shared \
		-t $(AWSBI) \
		output \
		M_AWS_ACCESS_KEY=$(AWS_ACCESS_KEY) \
		M_AWS_SECRET_KEY=$(AWS_SECRET_KEY)

destroy-plan: setup
	@docker run --rm \
		-v $(ROOT_DIR)/shared:/shared \
		-t $(AWSBI) \
		plan-destroy \
		M_AWS_ACCESS_KEY=$(AWS_ACCESS_KEY) \
		M_AWS_SECRET_KEY=$(AWS_SECRET_KEY)

destroy: setup
	@docker run --rm \
		-v $(ROOT_DIR)/shared:/shared \
		-t $(AWSBI) \
		destroy \
		M_AWS_ACCESS_KEY=$(AWS_ACCESS_KEY) \
		M_AWS_SECRET_KEY=$(AWS_SECRET_KEY)

setup: $(ROOT_DIR)/shared/$(VMS_RSA_FILENAME)

$(ROOT_DIR)/shared/$(VMS_RSA_FILENAME):
	@mkdir -p $(dir $@)
	@ssh-keygen -t rsa -b 4096 -f $@ -N '' 2>&1 >/dev/null
