package handler

import (
	"net/http"

	"github.com/rafaeljusto/dnsmanager"
	"github.com/rafaeljusto/dnsmanager/Godeps/_workspace/src/github.com/gustavo-hms/trama"
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
	return nil
}

func (d *savedDomain) Post(response trama.Response, r *http.Request) error {
	return nil
}

func (d *savedDomain) Templates() trama.TemplateGroupSet {
	return trama.NewTemplateGroupSet(templateFuncs)
}

func (d *savedDomain) Interceptors() trama.InterceptorChain {
	return defaultChain(d)
}
