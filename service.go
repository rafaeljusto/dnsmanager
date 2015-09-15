package dnsmanager

import (
	"net"
	"regexp"
	"strings"
)

const (
	dnsPort = 53
)

var (
	fqdnRX = regexp.MustCompile(`^((([a-z0-9][a-z0-9\-]*[a-z0-9])|[a-z0-9]+)\.)*([a-z]+|xn\-\-[a-z0-9]+)\.?$`)
)

type ServiceConfig struct {
	DNSServer struct {
		Address net.IP
		Port    int
		Zone    string
	} `yaml: "dns server"`
	BlockedTLDs  []string `yaml: "blocked TLDs"`
	DNSCheckPort int      `yaml: "dns check port"`
}

type Service interface {
	Save(Domain) error
	Retrieve(*TSigOptions) ([]Domain, error)
}

var NewService = func(config ServiceConfig) Service {
	config = setDefaults(config)
	return service{
		config: config,
	}
}

type service struct {
	config ServiceConfig
}

func (s service) Save(domain Domain) error {
	if err := s.validate(domain); err != nil {
		return err
	}

	if err := checkDelegation(domain, s.config.DNSCheckPort); err != nil {
		return err
	}

	return nsupdate(domain, s.config.DNSServer.Port)
}

func (s service) validate(domain Domain) error {
	var errBox ErrorBox

	domain.FQDN = strings.TrimSpace(domain.FQDN)
	domain.FQDN = strings.ToLower(domain.FQDN)
	domain.FQDN = strings.TrimRight(domain.FQDN, ".")

	if !fqdnRX.MatchString(domain.FQDN) {
		errBox.Append(NewGenericError(GenericErrorCodeInvalidFQDN))
	}

	for i, ns := range domain.Nameservers {
		ns.Hostname = strings.TrimSpace(ns.Hostname)
		ns.Hostname = strings.ToLower(ns.Hostname)
		ns.Hostname = strings.TrimRight(ns.Hostname, ".")
		domain.Nameservers[i] = ns

		if !fqdnRX.MatchString(ns.Hostname) {
			errBox.Append(NewDNSError(DNSErrorCodeInvalidFQDN, i, nil))
		}

		if ns.IPv4 != nil && ns.IPv4.To4() == nil {
			errBox.Append(NewDNSError(DNSErrorCodeInvalidIPv4Glue, i, nil))
		}
	}

	for i, ds := range domain.DSSet {
		ds.Digest = strings.Replace(ds.Digest, " ", "", -1)
		ds.Digest = strings.ToUpper(ds.Digest)
		domain.DSSet[i] = ds
	}

	labels := strings.Split(domain.FQDN, ".")
	tld := labels[len(labels)-1]
	for _, blockedTLD := range s.config.BlockedTLDs {
		if tld == blockedTLD {
			errBox.Append(NewGenericError(GenericErrorCodeBlockedTLD))
		}
	}

	return errBox.Unpack()
}

func (s service) Retrieve(tsig *TSigOptions) ([]Domain, error) {
	return axfr(
		s.config.DNSServer.Address,
		s.config.DNSServer.Port,
		s.config.DNSServer.Zone,
		tsig,
	)
}

func setDefaults(config ServiceConfig) ServiceConfig {
	if config.DNSServer.Address == nil {
		config.DNSServer.Address = net.ParseIP("127.0.0.1")
	}

	if config.DNSServer.Port == 0 {
		config.DNSServer.Port = dnsPort
	}

	if config.DNSServer.Zone == "" {
		config.DNSServer.Zone = "."
	}

	if config.DNSCheckPort == 0 {
		config.DNSCheckPort = dnsPort
	}

	return config
}
