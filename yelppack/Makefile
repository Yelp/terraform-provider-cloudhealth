# We need access to Bash features, such as arrays
SHELL := /bin/bash

.PHONY: itest_% clean shell
PROJECT = cloudhealth

VERSION = 4.4
ITERATION = yelp2
ARCH := $(shell facter architecture)

PACKAGE_NAME := terraform-provider-$(PROJECT)_$(VERSION)-$(ITERATION)_amd64.deb

DOCKER_TAG = terraform-provider-$(PROJECT)_$(shell date +%s)

CLEAN_CONTAINER := [ -e .docker_container_id ] && docker rm --force $$(cat .docker_container_id) || true; rm -f .docker_container_id

itest_%: dist/%/$(PACKAGE_NAME)
	docker run --rm -v $(CURDIR)/../dist:/dist:ro -v $(CURDIR)/itest.sh:/itest.sh:ro docker-dev.yelpcorp.com/$*_yelp:latest bash /itest.sh /dist/$*/$(PACKAGE_NAME) $(VERSION)

dist/%/$(PACKAGE_NAME): .docker_container_id
	mkdir -p ../dist/$* && \
	docker cp $$(cat .docker_container_id):/dist/$(PACKAGE_NAME) ../dist/$*/

.docker_container_id: .docker_image_id
	docker run --rm=false \
		-v $(CURDIR)/..:/go/src/terraform-provider-$(PROJECT):ro \
		-v $(CURDIR)/build.sh:/build.sh:ro \
		--cidfile=$(CURDIR)/.docker_container_id \
		$$(cat .docker_image_id) \
		bash /build.sh $(PROJECT) $(VERSION) $(ITERATION) $(TF_VERSIONS) || \
	(retval=$$?; $(CLEAN_CONTAINER); exit $$retval; )

.docker_image_id: Dockerfile go.mod go.sum
	docker build -t $(DOCKER_TAG) .
	docker inspect -f '{{ .Id }}' $(DOCKER_TAG) > .docker_image_id

clean:
	$(CLEAN_CONTAINER)
	rm .docker_image_id
	rm -rf ../dist

shell: .docker_image_id
	docker run --rm -t -i $$(cat .docker_image_id) /bin/bash
