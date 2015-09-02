#!/bin/sh
set -e

apt-get install bind9

mkdir -p /var/log/named
chown root:bind /var/log/named
chmod g+w /var/log/named

mkdir -p /etc/bind/zones
cp named.conf /etc/bind
chown -R root:bind /etc/bind
cp db.root /etc/bind/zones
chmod -R g+w /etc/bind/zones

service bind9 restart