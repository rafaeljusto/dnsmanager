package dnsmanager

import (
	"strings"

	"github.com/rafaeljusto/dnsmanager/Godeps/_workspace/src/github.com/miekg/dns"
)

func checkDelegation(domain Domain) error {
	for _, ns := range domain.Nameservers {
		checkNS(ns, domain.Name)
		checkDSs(domain.DSs, ns, domain.Name)
	}

	return nil
}

func checkNS(ns Nameserver, fqdn string) error {
	msg := new(dns.Msg)
	msg.SetQuestion(fqdn, dns.TypeSOA)
	msg.RecursionDesired = false

	server := ns.Name
	if strings.HasSuffix(ns.Name, fqdn) {
		server = ns.Glue
	}

	client := new(dns.Client)
	response, _, err := client.Exchange(msg, server+":53")

	if err == nil {
		if !response.Authoritative || len(response.Answer) == 0 {
			// TODO
		}

	} else {
		// TODO
	}

	return nil
}

func checkDSs(dsSet []DS, ns Nameserver, fqdn string) error {
	if len(ns.Name) == 0 {
		// Don't need to check DS if the nameserver is empty
		return nil
	}

	msg := new(dns.Msg)
	msg.SetQuestion(fqdn, dns.TypeDNSKEY)
	msg.RecursionDesired = false

	server := ns.Name
	if strings.HasSuffix(ns.Name, fqdn) {
		server = ns.Glue
	}

	client := new(dns.Client)
	response, _, err := client.Exchange(msg, server+":53")

	if err != nil {
		// TODO
	}

	if response.Truncated {
		client.Net = "tcp"
		response, _, err = client.Exchange(msg, server+":53")

		if err != nil {
			// TODO
		}
	}

	ds1 := dsSet[0]
	ds2 := dsSet[1]
	foundDS1 := false
	foundDS2 := false

	for _, rr := range response.Answer {
		if rr.Header().Rrtype == dns.TypeDNSKEY {
			dnskeyRR := rr.(*dns.DNSKEY)

			if int(dnskeyRR.KeyTag()) == ds1.KeyTag {
				foundDS1 = true

				ds := dnskeyRR.ToDS(ds1.DigestType)
				if int(ds.Algorithm) != ds1.Algorithm {
					// TODO
				} else if strings.ToUpper(ds.Digest) != ds1.Digest {
					// TODO
				} else if dnskeyRR.Flags&dns.SEP == 0 {
					// TODO
				}

			}

			if int(dnskeyRR.KeyTag()) == ds2.KeyTag {
				foundDS2 = true

				ds := dnskeyRR.ToDS(ds2.DigestType)
				if int(ds.Algorithm) != ds2.Algorithm {
					// TODO
				} else if strings.ToUpper(ds.Digest) != ds2.Digest {
					// TODO
				} else if dnskeyRR.Flags&dns.SEP == 0 {
					// TODO
				}

			}
		}
	}

	if !foundDS1 && len(ds1.Digest) > 0 {
		// TODO
	}

	if !foundDS2 && len(ds2.Digest) > 0 {
		// TODO
	}

	return nil
}
