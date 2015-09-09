package config

import (
	"net"

	"github.com/rafaeljusto/dnsmanager"
	"github.com/rafaeljusto/dnsmanager/Godeps/_workspace/src/gopkg.in/yaml.v2"
)

var WebNIC = struct {
	Home   string
	Log    string
	Listen struct {
		Address net.IP
		Port    int
	}
	TemplatePath string `yaml: "template path"`
	AssetsPath   string `yaml: "assets path"`
	TSig         dnsmanager.TSigOptions
	DNSManager   dnsmanager.ServiceConfig `yaml:"dns manager"`
}{}

func Load(file string) error {
	return yaml.Unmarshal([]byte(file), &WebNIC)
}
