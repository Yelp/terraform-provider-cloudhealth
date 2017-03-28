all: terraform-provider-cloudhealth

.PHONY: vendor
vendor:
	go get -u github.com/kardianos/govendor
	govendor sync

.PHONY: vet
vet:
	go tool vet *.go cloudhealth/*.go

.PHONY: test
test:
	go test ./cloudhealth

.PHONY: clean
clean:
	rm terraform-provider-cloudhealth

terraform-provider-cloudhealth: *.go cloudhealth/*.go
	go build

#
# Yelp-specific packaging
#
.PHONY: itest_%
itest_%:
	mkdir -p dist
	make -C yelppack $@

package: itest_lucid
