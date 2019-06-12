FROM docker-dev.yelpcorp.com/xenial_pkgbuild:latest

ENV PATH /usr/bin:/bin:/usr/sbin:/sbin:/usr/local/bin:/usr/local/sbin:/usr/local/go/bin:/go/bin
ENV GOPATH /go
RUN mkdir /go
#TODO: remove below
RUN gem install json  -v 1.8.3
RUN gem install ffi   -v 1.9.18
