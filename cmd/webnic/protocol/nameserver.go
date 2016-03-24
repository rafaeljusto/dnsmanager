package protocol

import (
	"net"

	"github.com/rafaeljusto/dnsmanager"
)

type Nameserver struct {
	Hostname string `json:"hostname"`
	IPv4     string `json:"ipv4"`
}

func NewNameserver(ns dnsmanager.Nameserver) Nameserver {
	n := Nameserver{
		Hostname: ns.Hostname,
	}

	if ns.IPv4 != nil {
		n.IPv4 = ns.IPv4.String()
	}

	return n
}

func (n Nameserver) Convert() dnsmanager.Nameserver {
	return dnsmanager.Nameserver{
		Hostname: n.Hostname,
		IPv4:     net.ParseIP(n.IPv4),
	}
}
