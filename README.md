# m-aws-basic-infrastructure

Epiphany Module: AWS Basic Infrastructure

# Basic usage

## Build image

In main directory run:

```shell
make build
```

## Run module with provided example

### Prepare config file

Prepare your own variables in vars.mk file to use in the building process.
Sample file (examples/basic_flow/vars.mk.sample):

```shell
AWS_ACCESS_KEY_ID = "xxx"
AWS_ACCESS_KEY_SECRET = "xxx"
```

### Create an environment

```shell
cd examples/basic_flow
make all
```

or step-by-step:

```shell
cd examples/basic_flow
make init
make plan
make apply
```

### Delete environment

```shell
cd examples/basic_flow
make all-destroy
```

or step-by-step

```shell
cd examples/basic_flow
make destroy-plan
make destroy
```

## Release module

```shell
make release
```

or if you want to set a different version number:

```shell
make release VERSION=number_of_your_choice
```
