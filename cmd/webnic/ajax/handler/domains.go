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
	Mux.Handle("/domains", func() handy.Handler { return &domains{} })
}

type domains struct {
	handy.DefaultHandler
	interceptor.IntrospectorCompliant

	Domains []dnsmanager.Domain `response:"get"`
}

func (d *domains) Get() int {
	service := dnsmanager.NewService(config.WebNIC.DNSManager)

	var err error
	d.Domains, err = service.Retrieve(&config.WebNIC.TSig)

	if err != nil {
		log.Println("error retrieving domains:", err)
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

func (d *domains) Interceptors() handy.InterceptorChain {
	return handy.NewInterceptorChain().
		Chain(interceptor.NewIntrospector(d)).
		Chain(interceptor.NewJSONCodec(d))
}
