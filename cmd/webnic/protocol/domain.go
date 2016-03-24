package protocol

import "github.com/rafaeljusto/dnsmanager"

type Domain struct {
	FQDN        string       `json:"fqdn"`
	Nameservers []Nameserver `json:"nameservers"`
	DSSet       []DS         `json:"dsset"`
}

func NewDomain(domain dnsmanager.Domain) Domain {
	d := Domain{
		FQDN: domain.FQDN,
	}

	for _, ns := range domain.Nameservers {
		d.Nameservers = append(d.Nameservers, NewNameserver(ns))
	}

	for _, ds := range domain.DSSet {
		d.DSSet = append(d.DSSet, NewDS(ds))
	}

	return d
}

func (d Domain) Convert() (domain dnsmanager.Domain, err error) {
	domain.FQDN = d.FQDN

	for _, ns := range d.Nameservers {
		domain.Nameservers = append(domain.Nameservers, ns.Convert())
	}

	var converted dnsmanager.DS

	for _, ds := range d.DSSet {
		if converted, err = ds.Convert(); err != nil {
			return
		}
		domain.DSSet = append(domain.DSSet, converted)
	}

	return
}
