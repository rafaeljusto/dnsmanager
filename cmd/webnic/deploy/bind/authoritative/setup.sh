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
dnssec-keygen -3 -f KSK -K /etc/bind/keys -r /dev/urandom .
dnssec-keygen -3 -K /etc/bind/keys -r /dev/urandom .

chown -R root:bind /etc/bind
chmod -R g+w /etc/bind/zones
chmod -R g+r /etc/bind/keys

# Give write permissions to /etc/bind
vi /etc/apparmor.d/usr.sbin.named

service apparmor restart
service bind9 restart
