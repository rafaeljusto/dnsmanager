package handler

import (
	"net/http"

	"github.com/rafaeljusto/dnsmanager"
	"github.com/rafaeljusto/dnsmanager/Godeps/_workspace/src/github.com/gustavo-hms/trama"
	handyinterceptor "github.com/rafaeljusto/dnsmanager/Godeps/_workspace/src/github.com/trajber/handy/interceptor"
	"github.com/rafaeljusto/dnsmanager/cmd/webnic/interceptor"
)

func init() {
	Mux.Register("/domain/{fqdn}", func() trama.Handler {
		return new(savedDomain)
	})
}

type savedDomain struct {
	trama.NopHandler
	FQDN   string            `urivar:"fqdn"`
	Domain dnsmanager.Domain `request:"post"`
}

func (d *savedDomain) Get(response trama.Response, r *http.Request) {

}

func (d *savedDomain) Post(response trama.Response, r *http.Request) {

}

func (d *savedDomain) Templates() trama.TemplateGroupSet {
	return trama.NewTemplateGroupSet(templateFuncs)
}

func (d *savedDomain) Interceptors() trama.InterceptorChain {
	return trama.NewInterceptorChain(
		handyinterceptor.NewURIVars(d),
		interceptor.NewPOST(d),
	)
}
