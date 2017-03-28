all: terraform-provider-cloudhealth

.PHONY: vendor
vendor:
	go get -u github.com/kardianos/govendor
	${GOPATH}/bin/govendor sync

.PHONY: vet
vet:
	go tool vet *.go cloudhealth/*.go

.PHONY: test
test:
	go test ./cloudhealth

.PHONY: clean
clean:
	rm -f terraform-provider-cloudhealth
	rm -rf dist/
	make -C yelppack clean

terraform-provider-cloudhealth: *.go cloudhealth/*.go vendor
	go build

#
# Yelp-specific packaging
#
.PHONY: itest_%
itest_%: vendor
	mkdir -p dist
	make -C yelppack $@

package: itest_lucid
