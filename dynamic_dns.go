package dnsmanager

import (
	"fmt"
	"time"

	"github.com/rafaeljusto/dnsmanager/Godeps/_workspace/src/github.com/miekg/dns"
)

func nsupdate(domain Domain, dnsPort int, tsigOptions *TSigOptions) error {
	var m dns.Msg
	m.SetUpdate(".")
	m.RemoveRRset([]dns.RR{
		&dns.NS{
			Hdr: dns.RR_Header{
				Name:   domain.FQDN,
				Rrtype: dns.TypeNS,
			},
		},
		&dns.A{
			Hdr: dns.RR_Header{
				Name:   domain.FQDN,
				Rrtype: dns.TypeA,
			},
		},
		&dns.DS{
			Hdr: dns.RR_Header{
				Name:   domain.FQDN,
				Rrtype: dns.TypeDS,
			},
		},
	})

	var newRRs []dns.RR
	for _, ns := range domain.Nameservers {
		rr, err := dns.NewRR(fmt.Sprintf("%s 172800 IN NS %s", domain.FQDN, ns.Hostname))
		if err != nil {
			return err
		}

		newRRs = append(newRRs, rr)

		if ns.IPv4 != nil {
			rr, err = dns.NewRR(fmt.Sprintf("%s 172800 IN A %s", domain.FQDN, ns.IPv4.String()))
			if err != nil {
				return err
			}

			newRRs = append(newRRs, rr)
		}
	}
	for _, ds := range domain.DSSet {
		rr, err := dns.NewRR(fmt.Sprintf("%s 172800 IN DS %d %d %d %s",
			domain.FQDN, ds.KeyTag, ds.Algorithm, ds.DigestType, ds.Digest))
		if err != nil {
			return err
		}

		newRRs = append(newRRs, rr)
	}

	m.Insert(newRRs)
	m.SetTsig(tsigOptions.Name, tsigOptions.Algorithm, 300, time.Now().Unix())

	var client dns.Client
	client.TsigSecret = map[string]string{
		tsigOptions.Name: tsigOptions.Secret,
	}

	_, _, err := client.Exchange(&m, "localhost:53")
	return err
}
