package handler

import (
	"net/http"

	"github.com/rafaeljusto/dnsmanager"
	"github.com/rafaeljusto/dnsmanager/Godeps/_workspace/src/github.com/gustavo-hms/trama"
	"github.com/rafaeljusto/dnsmanager/cmd/webnic/config"
)

func init() {
	Mux.Register("/domain", func() trama.Handler {
		return new(domain)
	})
}

type domain struct {
	defaultHandler
	Domain dnsmanager.Domain `request:"post"`
}

func (d *domain) Get(response trama.Response, r *http.Request) error {
	templateData := NewTemplateData()
	templateData.Action = "/domain"
	templateData.NewDomain = true

	// For now we are going to show only two nameservers
	for i := 0; i < 2; i++ {
		templateData.Domain.Nameservers = append(templateData.Domain.Nameservers, dnsmanager.Nameserver{})
	}

	// For now we are going to show only two DSs
	for i := 0; i < 2; i++ {
		templateData.Domain.DSSet = append(templateData.Domain.DSSet, dnsmanager.DS{})
	}

	var err error
	service := dnsmanager.NewService(config.WebNIC.DNSManager)
	templateData.RegisteredDomains, err = service.Retrieve(&config.WebNIC.TSig)
	if err != nil {
		return err
	}

	response.ExecuteTemplate("domain.html", &templateData)
	return nil
}

func (d *domain) Post(response trama.Response, r *http.Request) error {
	templateData := NewTemplateData()
	templateData.Domain = d.Domain

	service := dnsmanager.NewService(config.WebNIC.DNSManager)
	if err := service.Save(d.Domain); err != nil {
		return err
	}

	templateData.Success = true
	templateData.NewDomain = false
	templateData.Action = "/domain/" + d.Domain.FQDN

	var err error
	templateData.RegisteredDomains, err = service.Retrieve(&config.WebNIC.TSig)
	if err != nil {
		return err
	}

	response.ExecuteTemplate("domain.html", &templateData)
	return nil
}

func (d *domain) Templates() trama.TemplateGroupSet {
	return trama.NewTemplateGroupSet(templateFuncs)
}

func (d *domain) Interceptors() trama.InterceptorChain {
	return defaultChain(d)
}
