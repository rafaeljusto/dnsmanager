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
	Mux.Handle("/domains", func() handy.Handler { return &domains{} })
}

type domains struct {
	handy.DefaultHandler
	interceptor.IntrospectorCompliant

	Domains []protocol.Domain `response:"get"`
}

func (d *domains) Get() int {
	service := dnsmanager.NewService(config.WebNIC.DNSManager)
	domains, err := service.Retrieve(&config.WebNIC.TSig)

	if err != nil {
		log.Println("error retrieving domains:", err)
		return http.StatusInternalServerError
	}

	d.Domains = []protocol.Domain{}
	for _, domain := range domains {
		d.Domains = append(d.Domains, protocol.NewDomain(domain))
	}

	return http.StatusOK
}

func (d *domains) Interceptors() handy.InterceptorChain {
	return handy.NewInterceptorChain().
		Chain(interceptor.NewIntrospector(d)).
		Chain(interceptor.NewJSONCodec(d))
}
