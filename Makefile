all: terraform-provider-cloudhealth

.PHONY: vendor
vendor:
	go get ./...

.PHONY: vet
vet:
	go tool vet *.go cloudhealth/*.go

.PHONY: test
test: vendor
	go test ./cloudhealth

.PHONY: clean
clean:
	rm -f terraform-provider-cloudhealth
	rm -rf dist/
	make -C yelppack clean

terraform-provider-cloudhealth: vendor *.go cloudhealth/*.go
	go build

#
# Yelp-specific packaging
#
.PHONY: itest_%
itest_%:
	cp -vp go.mod go.sum yelppack/
	mkdir -p dist
	make -C yelppack $@

package: itest_bionic
