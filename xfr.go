package dnsmanager

import (
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/miekg/dns"
)

const (
	fudge = 300
)

type TSigOptions struct {
	Name      string
	Algorithm string
	Secret    string
}

func axfr(address net.IP, port int, zone string, tsigOptions *TSigOptions) ([]Domain, error) {
	client := new(dns.Client)
	client.Net = "tcp"

	if tsigOptions != nil {
		client.TsigSecret = map[string]string{
			tsigOptions.Name: tsigOptions.Secret,
		}
	}

	msg := new(dns.Msg)
	msg.SetAxfr(zone)

	if tsigOptions != nil {
		msg.SetTsig(tsigOptions.Name, tsigOptions.Algorithm, fudge, time.Now().Unix())
	}

	transfer := dns.Transfer{
		TsigSecret: client.TsigSecret,
	}

	addressPort := net.JoinHostPort(address.String(), strconv.Itoa(port))

	transferChannel, err := transfer.In(msg, addressPort)
	if err != nil {
		return nil, err
	}

	domains := make(map[string]Domain)
	glues := make(map[string]net.IP)

	for {
		response, ok := <-transferChannel
		if !ok {
			break
		}

		if response.Error != nil {
			return nil, response.Error
		}

		for _, rr := range response.RR {
			// ignore APEX records
			if rr.Header().Name == zone {
				continue
			}

			switch rr.Header().Rrtype {
			case dns.TypeNS:
				nsRR := rr.(*dns.NS)

				domain := domains[rr.Header().Name]
				domain.FQDN = rr.Header().Name
				domain.Nameservers = append(domain.Nameservers, Nameserver{
					Hostname: nsRR.Ns,
				})
				domains[rr.Header().Name] = domain

			case dns.TypeDS:
				dsRR := rr.(*dns.DS)

				domain := domains[rr.Header().Name]
				domain.FQDN = rr.Header().Name
				domain.DSSet = append(domain.DSSet, DS{
					KeyTag:     dsRR.KeyTag,
					Algorithm:  dsRR.Algorithm,
					DigestType: dsRR.DigestType,
					Digest:     strings.ToUpper(dsRR.Digest),
				})
				domains[rr.Header().Name] = domain

			case dns.TypeA:
				aRR := rr.(*dns.A)
				glues[aRR.Header().Name] = aRR.A
			}
		}
	}

	var result []Domain

	for _, domain := range domains {
		for i, nameserver := range domain.Nameservers {
			if glue, ok := glues[nameserver.Hostname]; ok {
				nameserver.IPv4 = glue
				domain.Nameservers[i] = nameserver
			}
		}

		result = append(result, domain)
	}

	return result, nil
}
