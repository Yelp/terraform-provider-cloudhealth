all: terraform-provider-cloudhealth

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

