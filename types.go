package dnsmanager

import "net"

type Domain struct {
	FQDN        string       `json:"fqdn"`
	Nameservers []Nameserver `json:"nameservers"`
	DSSet       []DS         `json:"dsset"`
}

type Nameserver struct {
	Hostname string `json:"hostname"`
	IPv4     net.IP `json:"ipv4"`
}

type DS struct {
	KeyTag     uint16 `json:"key-tag"`
	Algorithm  uint8  `json:"algorithm"`
	DigestType uint8  `json:"digest-type"`
	Digest     string `json:"digest"`
}
