package dnsmanager

import "net"

type Domain struct {
	FQDN        string       `domain:"fqdn"`
	Nameservers []Nameserver `domain:"nameservers"`
	DSSet       []DS         `domain:"dsset"`
}

type Nameserver struct {
	Name  string   `domain:"name"`
	Glues []net.IP `domain:"glues"`
}

type DS struct {
	KeyTag     uint16 `domain:"key-tag"`
	Algorithm  uint8  `domain:"algorithm"`
	DigestType uint8  `domain:"digest-type"`
	Digest     string `domain:"digest"`
}
