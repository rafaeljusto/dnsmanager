package dnsmanager

import (
	"net"
	"strconv"
	"strings"

	"github.com/miekg/dns"
)

func checkDelegation(domain Domain, dnsCheckPort int) error {
	for i, ns := range domain.Nameservers {
		if err := checkNS(ns, domain.FQDN, i, dnsCheckPort); err != nil {
			return err
		}

		if err := checkDSSet(domain.DSSet, ns, domain.FQDN, i, dnsCheckPort); err != nil {
			return err
		}
	}

	return nil
}

func checkNS(ns Nameserver, fqdn string, index int, dnsCheckPort int) error {
	response, err := query(fqdn, ns, index, dns.TypeSOA, dnsCheckPort)
	if err != nil {
		return err
	}

	if !response.Authoritative || len(response.Answer) == 0 {
		return NewDNSError(DNSErrorCodeNotAuthoritative, index, nil)
	}

	return nil
}

func checkDSSet(dsSet []DS, ns Nameserver, fqdn string, nsIndex int, dnsCheckPort int) error {
	response, err := query(fqdn, ns, nsIndex, dns.TypeDNSKEY, dnsCheckPort)
	if err != nil {
		return err
	}

	var errBox ErrorBox
	keytagMatch := make(map[uint16]bool)

	for _, rr := range response.Answer {
		if rr.Header().Rrtype != dns.TypeDNSKEY {
			continue
		}

		dnskeyRR := rr.(*dns.DNSKEY)

		for dsIndex, ds := range dsSet {
			if dnskeyRR.KeyTag() != ds.KeyTag {
				continue
			}

			keytagMatch[ds.KeyTag] = true
			dsFromResponse := dnskeyRR.ToDS(ds.DigestType)

			if dsFromResponse == nil {
				errBox.Append(NewDNSSECError(DNSSECErrorCodeInvalidDNSKEY, nsIndex, dsIndex))
				continue
			}

			if dsFromResponse.Algorithm != ds.Algorithm {
				errBox.Append(NewDNSSECError(DNSSECErrorCodeAlgorithmDontMatch, nsIndex, dsIndex))
			} else if strings.ToUpper(dsFromResponse.Digest) != ds.Digest {
				errBox.Append(NewDNSSECError(DNSSECErrorCodeDigestDontMatch, nsIndex, dsIndex))
			} else if dnskeyRR.Flags&dns.SEP == 0 {
				errBox.Append(NewDNSSECError(DNSSECErrorCodeDNSKEYNotSEP, nsIndex, dsIndex))
			}
		}
	}

	for dsIndex, ds := range dsSet {
		if _, ok := keytagMatch[ds.KeyTag]; !ok {
			errBox.Append(NewDNSSECError(DNSSECErrorCodeDSNotFound, nsIndex, dsIndex))
		}
	}

	return errBox.Unpack()
}

func query(fqdn string, ns Nameserver, index int, rrType uint16, dnsCheckPort int) (*dns.Msg, error) {
	msg := new(dns.Msg)
	msg.SetQuestion(fqdn, rrType)
	msg.RecursionDesired = false

	server := ns.Hostname
	if strings.HasSuffix(ns.Hostname, fqdn) {
		if ns.IPv4 == nil {
			return nil, NewDNSError(DNSErrorMissingGlue, index, nil)
		}

		server = ns.IPv4.String()
	}

	hostAndPort := net.JoinHostPort(server, strconv.Itoa(dnsCheckPort))

	client := new(dns.Client)
	response, _, err := client.Exchange(msg, hostAndPort)

	if err != nil {
		return nil, NewDNSError(DNSErrorCodeQueryFailed, index, err)
	}

	if response.Truncated {
		client.Net = "tcp"
		response, _, err = client.Exchange(msg, server+":53")

		if err != nil {
			return nil, NewDNSError(DNSErrorCodeQueryFailed, index, err)
		}
	}

	return response, nil
}
