FROM docker-dev.yelpcorp.com/trusty_yelp

RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -yq \
    wget \
    git \
    build-essential \
    ruby1.9.1 rubygems1.9.1 \
    libopenssl-ruby1.9.1 \
    ruby1.9.1-dev \
    --no-install-recommends

RUN wget https://storage.googleapis.com/golang/go1.8.3.linux-amd64.tar.gz && tar xzf go1.8.3.linux-amd64.tar.gz && mv go /usr/local
ENV PATH /usr/bin:/bin:/usr/sbin:/sbin:/usr/local/bin:/usr/local/sbin:/usr/local/go/bin:/go/bin
ENV GOPATH /go
RUN mkdir /go
#TODO: remove below
RUN gem install json  -v 1.8.3
RUN gem install fpm
