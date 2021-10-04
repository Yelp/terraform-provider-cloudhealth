FROM docker-dev.yelpcorp.com/bionic_pkgbuild

ENV PATH /usr/bin:/bin:/usr/sbin:/sbin:/usr/local/bin:/usr/local/sbin:/usr/local/go/bin:/go/bin
ENV GOPATH /go

WORKDIR /go/src/terraform-provider-cloudhealth
ADD go.mod go.sum ./
RUN go mod download
