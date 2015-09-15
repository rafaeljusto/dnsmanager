package handler

import (
	"log"
	"net/http"

	"github.com/rafaeljusto/dnsmanager"
	"github.com/rafaeljusto/dnsmanager/Godeps/_workspace/src/github.com/trajber/handy"
	"github.com/rafaeljusto/dnsmanager/Godeps/_workspace/src/github.com/trajber/handy/interceptor"
	"github.com/rafaeljusto/dnsmanager/cmd/webnic/config"
)

func init() {
	handle("/domain/{fqdn}", func() handy.Handler { return &domain{} })
}

type domain struct {
	handy.DefaultHandler
	interceptor.IntrospectorCompliant

	FQDN    string            `urivar:"fqdn"`
	Domain  dnsmanager.Domain `request:"put"`
	uriVars handy.URIVars
}

func (d *domain) URIVars() handy.URIVars {
	return d.uriVars
}

func (d *domain) Put() int {
	d.Domain.FQDN = d.FQDN

	service := dnsmanager.NewService(config.WebNIC.DNSManager)
	if err := service.Save(d.Domain); err != nil {
		log.Println("error saving domain:", err)
		return http.StatusInternalServerError
	}

	return http.StatusNoContent
}

func (d *domain) Interceptors() handy.InterceptorChain {
	return handy.NewInterceptorChain().
		Chain(interceptor.NewURIVars(d)).
		Chain(interceptor.NewJSONCodec(d))
}
