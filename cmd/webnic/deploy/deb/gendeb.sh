#!/bin/sh

usage() {
  echo "Usage: $1 <id>"
}

id="$1"

if [ -z "$id" ]; then
  echo "ID not defined!"
  usage $0
  exit 1
fi

pack_name="webnic"
version=`echo "$id" | awk -F "-" '{ print $1 }'`
release=`echo "$id" | awk -F "-" '{ print $2 }'`
maintainer="Rafael Dantas Justo <adm@rafael.net.br>"
url="http://github.com/rafaeljusto/dnsmanager"
license="MIT License"
description="Web NIC simulation"

if [ -z "$version" ]; then
  echo "Version not defined! Invalid ID"
  usage $0
  exit 1
fi

if [ -z "$release" ]; then
  echo "Release not defined! Invalid ID"
  usage $0
  exit 1
fi

install_path=/usr/local/webnic
tmp_dir=/tmp/webnic
project_root=$tmp_dir$install_path

workspace=`echo $GOPATH | cut -d: -f1`
workspace=$workspace/src/github.com/rafaeljusto/dnsmanager/cmd/webnic

# recompiling everything
current_dir=`pwd`
cd $workspace

shorthash=`git rev-parse --short --verify HEAD`
go build -ldflags "-X dnsmanager/cmd/webnic/config.Version=$version-$release-$shorthash"
cd $current_dir

if [ -f "${pack_name}_${id}_amd64.deb" ]; then
  # remove old deb
  rm "${pack_name}_${id}_amd64.deb"
fi

if [ -d $tmp_dir ]; then
  rm -rf $tmp_dir
fi

mkdir -p $project_root
mv $workspace/webnic $project_root/
cp -r $workspace/etc/conf $project_root/
cp -r $workspace/etc/templates $project_root/
cp -r $workspace/etc/assets $project_root/

fpm -s dir -t deb \
  --exclude=.git -n $pack_name -v "$version" --iteration "$release" \
  --maintainer "$maintainer" --url $url --license "$license" --description "$description" \
  --deb-upstart $workspace/deploy/deb/webnic.upstart \
  --deb-user root --deb-group root \
  --prefix / -C $tmp_dir usr/local/webnic
