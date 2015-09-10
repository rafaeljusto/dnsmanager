package handler

import (
	"net/http"
	"path"
	"strings"

	"github.com/rafaeljusto/dnsmanager"
	"github.com/rafaeljusto/dnsmanager/Godeps/_workspace/src/github.com/gustavo-hms/trama"
	"github.com/rafaeljusto/dnsmanager/cmd/webnic/config"
)

func init() {
	Mux.Register("/domain/{fqdn}", func() trama.Handler {
		return new(savedDomain)
	})
}

type savedDomain struct {
	defaultHandler
	FQDN   string            `urivar:"fqdn"`
	Domain dnsmanager.Domain `request:"post"`
}

func (d *savedDomain) Get(response trama.Response, r *http.Request) error {
	if !strings.HasSuffix(d.FQDN, ".") {
		d.FQDN += "."
	}

	templateData := NewTemplateData()
	templateData.Action = "/domain/" + d.Domain.FQDN

	var err error
	service := dnsmanager.NewService(config.WebNIC.DNSManager)
	templateData.RegisteredDomains, err = service.Retrieve(&config.WebNIC.TSig)
	if err != nil {
		return err
	}

	found := false
	for _, domain := range templateData.RegisteredDomains {
		if domain.FQDN == d.FQDN {
			templateData.Domain = domain
			found = true
			break
		}
	}

	if !found {
		// TODO!
	}

	response.ExecuteTemplate("domain.tmpl.html", &templateData)
	return nil
}

func (d *savedDomain) Post(response trama.Response, r *http.Request) error {
	service := dnsmanager.NewService(config.WebNIC.DNSManager)
	if err := service.Save(d.Domain); err != nil {
		return err
	}

	templateData := NewTemplateData()
	templateData.Domain = d.Domain
	templateData.Success = true
	templateData.Action = "/domain/" + d.Domain.FQDN

	var err error
	templateData.RegisteredDomains, err = service.Retrieve(&config.WebNIC.TSig)
	if err != nil {
		return err
	}

	response.ExecuteTemplate("domain.tmpl.html", &templateData)
	return nil
}

func (d *savedDomain) Templates() trama.TemplateGroupSet {
	groupSet := trama.NewTemplateGroupSet(templateFuncs)

	for _, language := range config.WebNIC.Templates.Languages {
		templatePath := path.Join(config.WebNIC.Home, config.WebNIC.Templates.Path, language, "domain.tmpl.html")
		groupSet.Insert(trama.TemplateGroup{
			Name:  language,
			Files: []string{templatePath},
		})
	}

	return groupSet
}

func (d *savedDomain) Interceptors() trama.InterceptorChain {
	return defaultChain(d)
}
