#!/bin/sh
set -e

apt-get install bind9 ntp

mkdir -p /var/log/named
chown root:bind /var/log/named
chmod g+w /var/log/named

cp named.conf /etc/bind
chown -R root:bind /etc/bind
cp db.root /etc/bind

service bind9 restart