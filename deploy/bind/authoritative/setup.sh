#!/bin/sh
set -e

apt-get install bind9 ntp

mkdir -p /var/log/named
chown root:bind /var/log/named
chmod g+w /var/log/named

cp named.conf /etc/bind

mkdir -p /etc/bind/zones
cp db.root /etc/bind/zones

mkdir -p /etc/bind/keys
cd /etc/bind/keys
dnssec-keygen -3 -f KSK -r /dev/urandom .

cd /etc/bind/zones
dnssec-signzone -S -o . -3 abc123 -K /etc/bind/keys -z db.root

chown -R root:bind /etc/bind
chmod -R g+w /etc/bind/zones
chmod -R g+r /etc/bind/keys

service bind9 restart