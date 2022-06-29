#!/bin/bash
set -e

project=$1 ; shift
version=$1 ; shift
iteration=$1 ; shift

go build -v -o /go/bin/terraform-provider-${project}

mkdir /dist && cd /dist

install_path="/usr/local/share/terraform/plugins/terraform-registry.yelpcorp.com/yelp/${project}/${version}/linux_amd64/"
echo "Install path is ${install_path}"

fpm -s dir -t deb --deb-no-default-config-files --name terraform-provider-${project} \
  --iteration ${iteration} --version ${version} \
  /go/bin/terraform-provider-${project}=${install_path}

ls /dist
