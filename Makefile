# HELP
# This will output the help for each task
# thanks to https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help

help: ## This help.
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help

IMAGE_NAME := "cert-manager-webhook-oci"
IMAGE_TAG := "latest"

OUT := $(shell pwd)/deploy

$(shell mkdir -p "$(OUT)")

.PHONY: verify
verify: ## verify the code
	go test -v .

.PHONY: build
build: ## build the docker image
	docker build -t "$(IMAGE_NAME):$(IMAGE_TAG)" .

.PHONY: manifest
manifest: ## create the single manifest YAML distribution
	VERSION=$(shell cat deploy/cert-manager-webhook-oci/Chart.yaml | grep version | awk -F":" '{print $2}' | tr -d 'version:" '); \
	mkdir -p $(OUT)/v$${VERSION}; \
	helm template \
	    cert-manager-webhook-oci \
		--namespace cert-manager \
        deploy/cert-manager-webhook-oci > "$(OUT)/v$${VERSION}/cert-manager-webhook-oci.yaml"
