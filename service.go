package dnsmanager

import "net"

const (
	dnsPort = 53
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
	return nil
}

type service struct {
	config ServiceConfig
}

func (s service) Save(domain Domain) error {
	// TODO: Check blocked TLDs
	if err := checkDelegation(domain); err != nil {
		return err
	}

	return nsupdate(domain, s.config.DNSServer.Port)
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
