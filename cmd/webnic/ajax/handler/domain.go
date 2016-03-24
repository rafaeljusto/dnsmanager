package handler

import (
	"log"
	"net/http"

	"github.com/rafaeljusto/dnsmanager"
	"github.com/rafaeljusto/dnsmanager/cmd/webnic/config"
	"github.com/rafaeljusto/dnsmanager/cmd/webnic/protocol"
	"github.com/trajber/handy"
	"github.com/trajber/handy/interceptor"
)

func init() {
	Mux.Handle("/domain/{fqdn}", func() handy.Handler { return &domain{} })
}

type domain struct {
	handy.DefaultHandler
	interceptor.IntrospectorCompliant

	FQDN   string          `urivar:"fqdn"`
	Domain protocol.Domain `request:"put"`
}

func (d *domain) Put() int {
	d.Domain.FQDN = d.FQDN

	domain, err := d.Domain.Convert()
	if err != nil {
		// TODO: BadRequest body
		return http.StatusBadRequest
	}

	service := dnsmanager.NewService(config.WebNIC.DNSManager)
	if err := service.Save(domain, &config.WebNIC.TSig); err != nil {
		log.Print("error saving domain:", err)
		return http.StatusInternalServerError
	}

	return http.StatusNoContent
}

func (d *domain) Interceptors() handy.InterceptorChain {
	return handy.NewInterceptorChain().
		Chain(interceptor.NewIntrospector(d)).
		Chain(interceptor.NewURIVars(d)).
		Chain(interceptor.NewJSONCodec(d))
}
