package config

import (
	"io/ioutil"
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
	Templates struct {
		Path      string
		Languages []string
	}
	AssetsPath string `yaml: "assets path"`
	TSig       dnsmanager.TSigOptions
	DNSManager dnsmanager.ServiceConfig `yaml:"dns manager"`
}{}

func Load(file string) error {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, &WebNIC)
}
