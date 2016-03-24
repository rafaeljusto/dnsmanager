package dnsmanager

import "net"

type Domain struct {
	FQDN        string
	Nameservers []Nameserver
	DSSet       []DS
}

type Nameserver struct {
	Hostname string
	IPv4     net.IP
}

type DS struct {
	KeyTag     uint16
	Algorithm  uint8
	DigestType uint8
	Digest     string
}
