#!/bin/bash

set -eu

package=$1
version=$2

install_path="/usr/local/share/terraform/plugins/terraform-registry.yelpcorp.com/yelp/cloudhealth/${version}/linux_amd64/terraform-provider-cloudhealth"

dpkg -i "$package"
ls -la $install_path
test -x $install_path
